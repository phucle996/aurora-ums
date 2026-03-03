package app

import (
	etcdinfra "aurora/infra/etcd"
	redisinfra "aurora/infra/redis"
	"aurora/internal/app/txmanager"
	"aurora/internal/cache"
	"aurora/internal/config"
	domainrepo "aurora/internal/domain/repository"
	domainsvc "aurora/internal/domain/service"
	"aurora/internal/ratelimit"
	repoImple "aurora/internal/repository"
	svcImple "aurora/internal/service"
	"aurora/internal/transport/http/handler"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Modules struct {
	// Infrastructure
	PostgresDB *pgxpool.Pool
	Redis      *redis.Client
	Etcd       *clientv3.Client

	AuthHandler  *handler.AuthHandler
	UserHandler  *handler.UserHandler
	MfaHandler   *handler.MfaHandler
	RbacHandler  *handler.RbacHandler
	RateLimiter  *ratelimit.Bucket
	Publisher    redisinfra.EventPublisher
	Token        *config.TokenCfg
	DeviceSvc    domainsvc.DeviceSvcInterface
	DeviceRepo   domainrepo.DeviceRepoInterface
	UserRepo     domainrepo.UserRepoInterface
	RbacRepo     domainrepo.RBACRepoInterface
	JwtBlacklist *cache.JWTBlacklist
	DeviceCache  *cache.DeviceSecretCache
	PermCache    *cache.UserPermissionCache
	MFASession   *cache.MFASessionCache
}

// NewModules assembles all infrastructure dependencies.
func NewModules(
	ctx context.Context,
	db *pgxpool.Pool,
	redis *redis.Client,
	publisher redisinfra.EventPublisher,
	cfg *config.Config,
) (*Modules, error) {
	etcdClient, err := etcdinfra.NewClient(&cfg.Etcd)
	if err != nil {
		return nil, err
	}

	txMgr := txmanager.NewPgxTxManager(db)
	rlBucket := ratelimit.NewBucket(redis)
	jwtBlacklist := cache.NewJWTBlacklist(redis)
	deviceCache := cache.NewDeviceSecretCache(redis)
	permCache := cache.NewUserPermissionCache(redis)
	mfaSession := cache.NewMFASessionCache(redis)

	userRepo := repoImple.NewUserRepoImple(db)
	ottrepo := repoImple.NewOttRepoImple(db)
	refreshRepo := repoImple.NewRefreshRepoImple(db)
	deviceRepo := repoImple.NewDeviceRepoImple(db)
	mfaRepo := repoImple.NewMFARepoImple(db)
	rbacRepo := repoImple.NewRBACRepoImple(db)

	ottSvc := svcImple.NewOttSvcImple(&cfg.Token, ottrepo)
	refreshSvc := svcImple.NewRefreshSvcImple(refreshRepo, &cfg.Token)
	deviceSvc := svcImple.NewDeviceSvcImple(deviceRepo)
	mfaSvc := svcImple.NewMFASvcImple(mfaRepo)
	rbacSvc := svcImple.NewRBACSvcImple(rbacRepo, permCache)
	userSvc := svcImple.NewUserSvcImple(userRepo, ottSvc, rbacSvc, publisher, txMgr)
	authSvc := svcImple.NewAuthSvcImple(userSvc, ottSvc, refreshSvc, deviceSvc, rbacSvc, mfaSvc, publisher, &cfg.Token, jwtBlacklist, deviceCache, permCache, mfaSession)

	authHandler := handler.NewAuthHandler(authSvc, &cfg.Token)
	userHandler := handler.NewUserHandler(userSvc)
	mfaHandler := handler.NewMfaHandler(mfaSvc)
	rbacHandler := handler.NewRbacHandler(rbacSvc)

	return &Modules{
		AuthHandler:  authHandler,
		UserHandler:  userHandler,
		MfaHandler:   mfaHandler,
		RbacHandler:  rbacHandler,
		PostgresDB:   db,
		Redis:        redis,
		Etcd:         etcdClient,
		RateLimiter:  rlBucket,
		Publisher:    publisher,
		Token:        &cfg.Token,
		DeviceSvc:    deviceSvc,
		DeviceRepo:   deviceRepo,
		UserRepo:     userRepo,
		RbacRepo:     rbacRepo,
		JwtBlacklist: jwtBlacklist,
		DeviceCache:  deviceCache,
		PermCache:    permCache,
		MFASession:   mfaSession,
	}, nil
}
