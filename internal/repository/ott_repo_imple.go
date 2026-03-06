package repoImple

import (
	"aurora/internal/domain/entity"
	domainrepo "aurora/internal/domain/repository"
	"aurora/internal/errorx"
	"context"
	"crypto/subtle"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

type OttRepoImple struct {
	redis         *redis.Client
	purposePrefix string
}

var consumeOTTScript = redis.NewScript(`
local key = KEYS[1]
local expected = ARGV[1]
local current = redis.call("GET", key)
if not current then
  return 0
end
if current == expected then
  redis.call("DEL", key)
  return 1
end
return -1
`)

func NewOttRepoImple(redisClient *redis.Client) domainrepo.OttRepoInterface {
	return &OttRepoImple{
		redis:         redisClient,
		purposePrefix: "ott:user-purpose:",
	}
}

func (r *OttRepoImple) Create(ctx context.Context, ott *entity.OneTimeToken) error {
	userID, purpose, tokenHash, err := validateOTTInput(ott)
	if err != nil {
		return err
	}
	if r == nil || r.redis == nil {
		return errors.New("redis client is nil")
	}

	ttl := time.Until(ott.ExpiresAt)
	if ttl <= 0 {
		return errorx.ErrInvalidArgument
	}

	key := r.purposeKey(userID, purpose)
	return r.redis.Set(ctx, key, tokenHash, ttl).Err()
}

func (r *OttRepoImple) Validate(ctx context.Context, ott *entity.OneTimeToken) error {
	userID, purpose, tokenHash, err := validateOTTInput(ott)
	if err != nil {
		return err
	}
	if r == nil || r.redis == nil {
		return errors.New("redis client is nil")
	}

	raw, err := r.redis.Get(ctx, r.purposeKey(userID, purpose)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return errorx.ErrOttNotFound
		}
		return err
	}

	if subtle.ConstantTimeCompare([]byte(strings.TrimSpace(raw)), []byte(tokenHash)) != 1 {
		return errorx.ErrOttNotFound
	}
	return nil
}

func (r *OttRepoImple) ConsumTx(ctx context.Context, tx pgx.Tx, ott *entity.OneTimeToken) error {
	_ = tx
	return r.consume(ctx, ott)
}

func (r *OttRepoImple) consume(ctx context.Context, ott *entity.OneTimeToken) error {
	userID, purpose, tokenHash, err := validateOTTInput(ott)
	if err != nil {
		return err
	}
	if r == nil || r.redis == nil {
		return errors.New("redis client is nil")
	}

	result, err := consumeOTTScript.Run(
		ctx,
		r.redis,
		[]string{r.purposeKey(userID, purpose)},
		tokenHash,
	).Int()
	if err != nil {
		return err
	}

	switch result {
	case 1:
		return nil
	case 0, -1:
		return errorx.ErrOttNotFound
	default:
		return errorx.ErrOttNotFound
	}
}

func (r *OttRepoImple) purposeKey(userID, purpose string) string {
	return r.purposePrefix + strings.TrimSpace(userID) + ":" + strings.TrimSpace(purpose)
}

func validateOTTInput(ott *entity.OneTimeToken) (string, string, string, error) {
	if ott == nil {
		return "", "", "", errorx.ErrEntityNil
	}

	userID := ott.UserID.String()
	tokenHash := strings.TrimSpace(ott.TokenHash)
	purpose := strings.TrimSpace(string(ott.Purpose))
	if ott.UserID == uuid.Nil || tokenHash == "" || purpose == "" {
		return "", "", "", errorx.ErrInvalidArgument
	}
	return userID, purpose, tokenHash, nil
}
