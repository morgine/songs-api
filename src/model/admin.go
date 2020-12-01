package model
//
//import (
//	"github.com/morgine/songs/src/message"
//	"golang.org/x/crypto/bcrypt"
//)
//
//type Admin struct {
//	ID       int
//	Username string `gorm:"uniqueIndex;comment:用户名"`
//	Password string `gorm:"comment:密码"`
//	Avatar   string
//}
//
//type AdminGorm struct {
//	g *Gorm
//}
//
//func (g *Gorm) Admin() *AdminGorm {
//	return &AdminGorm{g: g}
//}
//
//func (g *AdminGorm) RegisterAdmin(username, password string) (err error) {
//	admin, err := g.GetAdmin(username)
//	if err != nil {
//		return err
//	}
//	if admin != nil {
//		return message.ErrAdminUsernameAlreadyExist
//	} else {
//		password, err := bcrypt.GenerateFromPassword([]byte(password), 10)
//		if err != nil {
//			return err
//		}
//		return g.g.Create(&Admin{Username: username, Password: string(password)})
//	}
//}
//
//func (g *AdminGorm) LoginAdmin(username, password string) (*Admin, error) {
//	admin, err := g.GetAdmin(username)
//	if err != nil {
//		return nil, err
//	}
//	if admin == nil {
//		return nil, message.ErrAdminUsernameOrPasswordIncorrect
//	} else {
//		err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password))
//		if err != nil {
//			if err == bcrypt.ErrMismatchedHashAndPassword {
//				return nil, message.ErrAdminUsernameOrPasswordIncorrect
//			} else {
//				return nil, err
//			}
//		} else {
//			return admin, nil
//		}
//	}
//}
//
//func (g *AdminGorm) GetAdmin(username string) (*Admin, error) {
//	admin := &Admin{}
//	err := g.g.First(admin, Where("username=?", username))
//	if err != nil {
//		return nil, err
//	}
//	if admin.ID > 0 {
//		return admin, nil
//	} else {
//		return nil, nil
//	}
//}
//
//func (g *AdminGorm) ResetPassword(authAdminID int, newPassword string) error {
//	password, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
//	if err != nil {
//		return err
//	}
//	return g.g.Updates(&Admin{Password: string(password)}, Where("id=?", authAdminID))
//}
