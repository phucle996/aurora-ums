package svcImple

import (
	redisinfra "aurora/infra/redis"
	"aurora/internal/app/txmanager"
	"aurora/internal/domain/entity"
	domainrepo "aurora/internal/domain/repository"
	domainsvc "aurora/internal/domain/service"
	"aurora/internal/errorx"
	"aurora/internal/security"
	"context"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserSvcImple struct {
	UserRepo  domainrepo.UserRepoInterface
	OttSvc    domainsvc.OttSvcInterface
	RbacSvc   domainsvc.RBACSvcInterface
	Publisher redisinfra.EventPublisher
	txMgr     txmanager.TxManager
}

func NewUserSvcImple(
	UserRepo domainrepo.UserRepoInterface,
	OttSvc domainsvc.OttSvcInterface,
	RbacSvc domainsvc.RBACSvcInterface,
	Publisher redisinfra.EventPublisher,
	txMgr txmanager.TxManager,
) domainsvc.UserSvcInterface {
	return &UserSvcImple{
		UserRepo:  UserRepo,
		OttSvc:    OttSvc,
		RbacSvc:   RbacSvc,
		Publisher: Publisher,
		txMgr:     txMgr,
	}

}

func (s *UserSvcImple) CreateUser(ctx context.Context, user *entity.User) error {
	// check user exist
	if err := s.UserRepo.CheckUserExist(ctx, user.Username, user.Email); err != nil {
		return err
	}

	// hash password
	passwdHash, err := security.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = passwdHash

	// call repo for persistant user to db
	if err := s.UserRepo.CreateUser(ctx, user); err != nil {
		return err
	}
	return nil
}

func (s *UserSvcImple) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	return s.UserRepo.GetUserByUsername(ctx, username)
}

func (s *UserSvcImple) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	return s.UserRepo.GetUserByEmail(ctx, email)
}

func (s *UserSvcImple) GetUserWithProfileByID(ctx context.Context, userID uuid.UUID) (*entity.UserWithProfile, error) {
	return s.UserRepo.GetUserWithProfileByID(ctx, userID)
}

func (s *UserSvcImple) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*entity.CurrentUserView, error) {
	userWithProfile, err := s.UserRepo.GetUserWithProfileByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	roleNames := make([]string, 0)
	permissions := make([]string, 0)
	if s.RbacSvc != nil {
		roles, err := s.RbacSvc.ListUserRoles(ctx, userID)
		if err != nil {
			return nil, err
		}
		roleNames = make([]string, 0, len(roles))
		permSet := make(map[string]struct{})
		for _, role := range roles {
			roleNames = append(roleNames, role.Name)

			rolePerms, rolePermErr := s.RbacSvc.ListRolePermissions(ctx, role.ID)
			if rolePermErr != nil {
				return nil, rolePermErr
			}
			for _, perm := range rolePerms {
				permName := strings.TrimSpace(perm.Name)
				if permName == "" {
					continue
				}
				permSet[permName] = struct{}{}
			}
		}

		permissions = make([]string, 0, len(permSet))
		for name := range permSet {
			permissions = append(permissions, name)
		}
		sort.Strings(permissions)
	}

	return &entity.CurrentUserView{
		User:        userWithProfile.User,
		Profile:     userWithProfile.Profile,
		Roles:       roleNames,
		Permissions: permissions,
	}, nil
}

func (s *UserSvcImple) UpsertProfile(ctx context.Context, profile *entity.Profile) error {
	return s.UserRepo.UpsertProfile(ctx, profile)
}

func (s *UserSvcImple) UpdateUserStateTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, status entity.UserStatus) error {
	return s.UserRepo.UpdateStatusUserTx(ctx, tx, userID, status)
}

func (s *UserSvcImple) GetUserStatusTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (entity.UserStatus, error) {
	return s.UserRepo.GetUserStatusTx(ctx, tx, userID)
}

func (s *UserSvcImple) UpdatePasswordTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, passwordHash string) error {
	return s.UserRepo.UpdatePasswordTx(ctx, tx, userID, passwordHash)
}

func (s *UserSvcImple) ActiveAccount(ctx context.Context, userID uuid.UUID, token string) error {
	return s.txMgr.WithTx(ctx, func(tx pgx.Tx) error {
		status, err := s.UserRepo.GetUserStatusTx(ctx, tx, userID)
		if err != nil {
			return err
		}
		if status == entity.UserStatusActive {
			return errorx.ErrAccountAlreadyActivated
		}

		if err := s.OttSvc.ConsumTokenTx(ctx, tx, userID, token, entity.OttAccountVerify); err != nil {
			return err
		}

		if err := s.UserRepo.UpdateStatusUserTx(ctx, tx, userID, entity.UserStatusActive); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserSvcImple) VerifyResetPassword(ctx context.Context, userID uuid.UUID, token string) error {
	return s.OttSvc.ValidateToken(ctx, userID, token, entity.OttPasswordReset)
}

func (s *UserSvcImple) NewPassword(ctx context.Context, userID uuid.UUID, token, password string) error {
	return s.txMgr.WithTx(ctx, func(tx pgx.Tx) error {
		if err := s.OttSvc.ConsumTokenTx(ctx, tx, userID, token, entity.OttPasswordReset); err != nil {
			return err
		}

		passwdHash, err := security.HashPassword(password)
		if err != nil {
			return err
		}

		if err := s.UserRepo.UpdatePasswordTx(ctx, tx, userID, passwdHash); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserSvcImple) ForgotPasswd(ctx context.Context, email string) error {
	user, err := s.UserRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	plainToken, err := s.OttSvc.CreateToken(ctx, user.ID, entity.OttPasswordReset)
	if err != nil {
		return err
	}
	if err := s.Publisher.Publish(ctx, "auth.forgot_passwd", map[string]any{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"token":   plainToken,
	}); err != nil {
		return err
	}

	return nil
}
