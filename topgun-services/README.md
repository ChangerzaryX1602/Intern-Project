# Coding style Example

TOP GUN SERVICES

## Response Format

```shell
            return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Invalid filter",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
```

## Always test on complex services (You can test on other methods ex.Unit test, Integration test, etc.)
### touch {module}_test.go
```shell
package user_test

import (
	"errors"
	"log"
	"testing"
	"time"

	"github.com/Zentrix-Software-Hive/zyntax-ai-services/pkg/models"
	"github.com/Zentrix-Software-Hive/zyntax-ai-services/pkg/user"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Test struct {
	TestName string
	Func     func() error
}

func TestUser(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	userRepository := user.NewUserRepository(db)
	userService := user.NewUserService(userRepository)

	tests := []Test{
		{
			TestName: "CreateUser",
			Func: func() error {
				u := models.User{
					Username: "testuser",
					Password: "testpassword",
					FullName: "Test User",
					Address:  "123 Test St",
					Token:    0,
				}
				_, err := userService.CreateUser(u)
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			TestName: "GetUser",
			Func: func() error {
				_, err := userService.GetUser(1)
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			TestName: "UpdateUser",
			Func: func() error {
				_, err := userService.UpdateUser(1, models.User{
					Username: "updateduser",
					Password: "updatedpassword",
				})
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			TestName: "DeleteUser",
			Func: func() error {
				err := userService.DeleteUser(1)
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			TestName: "ReduceToken",
			Func: func() error {
				newUser := models.User{
					Username:  "tokentest",
					Password:  "pass",
					FullName:  "Token Test",
					Address:   "456 Token Ave",
					Token:     10,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				createdUser, err := userService.CreateUser(newUser)
				if err != nil {
					return err
				}
				updatedUser, err := userService.ReduceToken(createdUser.ID, 5)
				if err != nil {
					return err
				}
				if updatedUser.Token != 5 {
					return errors.New("token reduction did not work as expected")
				}
				return nil
			},
		},
		{
			TestName: "GetUserAfterDelete",
			Func: func() error {
				_, err := userService.GetUser(1)
				if err == nil {
					return errors.New("expected error when getting deleted user")
				}
				return nil
			},
		},
		{
			TestName: "CreateDuplicateUser",
			Func: func() error {
				u := models.User{
					Username: "dupuser",
					Password: "password",
					FullName: "Duplicate User",
					Address:  "456 Dup St",
				}
				_, err := userService.CreateUser(u)
				if err != nil {
					return err
				}
				_, err = userService.CreateUser(u)
				if err == nil {
					return errors.New("expected error when creating duplicate user, got nil")
				}
				return nil
			},
		},
		{
			TestName: "GetNonExistentUser",
			Func: func() error {
				_, err := userService.GetUser(999)
				if err == nil {
					return errors.New("expected error when fetching non-existent user, got nil")
				}
				return nil
			},
		},
		{
			TestName: "UpdateNonExistentUser",
			Func: func() error {
				_, err := userService.UpdateUser(999, models.User{Username: "nonexistent"})
				if err == nil {
					return errors.New("expected error when updating non-existent user, got nil")
				}
				return nil
			},
		},
		{
			TestName: "DeleteNonExistentUser",
			Func: func() error {
				err := userService.DeleteUser(999)
				if err == nil {
					return errors.New("expected error when deleting non-existent user, got nil")
				}
				return nil
			},
		},
		{
			TestName: "ReduceTokenInsufficient",
			Func: func() error {
				newUser := models.User{
					Username:  "tokenUser",
					Password:  "pass",
					FullName:  "Token User",
					Address:   "789 Token Road",
					Token:     3,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				createdUser, err := userService.CreateUser(newUser)
				if err != nil {
					return err
				}
				_, err = userService.ReduceToken(createdUser.ID, 5)
				if err == nil {
					return errors.New("expected error when reducing tokens with insufficient balance, got nil")
				}
				return nil
			},
		},
		{
			TestName: "CreateUserEmptyUsername",
			Func: func() error {
				u := models.User{
					Username: "",
					Password: "password",
					FullName: "Empty Username",
					Address:  "No Address",
					Token:    0,
				}
				_, err := userService.CreateUser(u)
				if err == nil {
					return errors.New("expected error when creating user with empty username, got nil")
				}
				return nil
			},
		},
	}

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			if err := test.Func(); err != nil {
				t.Errorf("Test %s failed with error: %v", test.TestName, err)
			}
		})
	}
}
```
## Swagger Ignore
### swaggerignore:"true"
```shell
	Role         string    `json:"role" gorm:"enum('staff', 'applicant');default:'applicant'" swaggerignore:"true"`
``` 
## Primary Key
### If you want to add a auto increment primary key, Always use integer!
```shell
    ID uint `json:"id" gorm:"primaryKey;autoIncrement"`
```

## Soft Delete
```shell
    DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
```

## Datetime
```shell
    CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;default:CURRENT_TIMESTAMP"`
    UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime;default:CURRENT_TIMESTAMP"`
```
## Router
### Do not write the endpoint path in infrastructure folder write it on the handler file
```shell
func NewAuthHandler(router fiber.Router, authService domain.AuthService, userService domain.UserService) {
	handler := &authHandler{
		authService: authService,
		userService: userService,
	}
	router.Post("/register", handler.Register())
	router.Post("/login", handler.Login())
}
```
## Always use !fiber.IsChild on cron job function (Handle the chopsticks problems)
```shell
    if !fiber.IsChild() {
        // Implement your cron job logic here
        return
    }
```

## Always use pagination and filter on service that's return a list of data
```shell
func (r *userRepository) GetUsers(pagination models.Pagination, filter models.Search) ([]models.User, *models.Pagination, *models.Search, error) {
	if r.DB == nil {
		return nil, nil, nil, gorm.ErrInvalidDB
	}
	var users []models.User
	dbTx := r.DB
	if filter.Keyword != "" || filter.Column != "" {
		dbTx = utils.ApplySearch(r.DB, filter)
	}
	var total int64
	err := dbTx.Model(&models.User{}).Count(&total).Error
	if err != nil {
		return nil, nil, nil, err
	}
	pagination.Total = total
	err = dbTx.Offset((pagination.Page - 1) * pagination.PerPage).
		Limit(pagination.PerPage).
		Find(&users).Error
	if err != nil {
		return nil, nil, nil, err
	}
	return users, &pagination, &filter, nil
}
```

## Always use jwt middleware on the handler that's require permission
### routerResource.ReqAuthHandler()
```shell
func NewUserHandler(router fiber.Router, routerResource *handlers.RouterResources, userService domain.UserService) {
	handler := &userHandler{
		userService: userService,
	}
	router.Get("/", routerResource.ReqAuthHandler(), handler.GetUsers())
	router.Get("/:id", routerResource.ReqAuthHandler(), handler.GetUser())
	router.Post("/", routerResource.ReqAuthHandler(), handler.CreateUser())
	router.Put("/:id", routerResource.ReqAuthHandler(), handler.UpdateUser())
	router.Delete("/:id", routerResource.ReqAuthHandler(), handler.DeleteUser())
}
``` 
## Always assign the meaningful name to the variable
### Don't do this
```shell
x := models.User{}
```
### Do this
```shell
user := models.User{}
```
## If you want to create a new file, always check if the directory is exist
```shell
package utils
import (
    "os"
    "path/filepath"
)
func CreateDirectoryIfNotExists(dir string) error {
    if _, err := os.Stat(dir); os.IsNotExist(err) {
        err = os.MkdirAll(dir, os.ModePerm)
        if err != nil {
            return err
        }
    }
    return nil
}
```
# IMPORTANT!!!
## Repository
English: Used for communication with external services such as Database, API, File System, Cache, etc.

Thai: ใช้สำหรับการสื่อสารกับบริการภายนอก เช่น ฐานข้อมูล, API, ระบบไฟล์, แคช เป็นต้น
## Service (Usecase)
### Always reuse the repository function here!!
English: Used for business logic, data processing, and data validation.

Thai: ใช้สำหรับตรรกะทางธุรกิจ, การประมวลผลข้อมูล, และการตรวจสอบข้อมูล
## Handler (Controller)
### Always reuse the service function here!!
English: Used for communication with the client, such as HTTP request and response.

Thai: ใช้สำหรับการสื่อสารกับลูกค้า เช่น คำขอ HTTP และการตอบสนอง

## Cycle import and circular dependency 
## Do not do this
```shell
package user
import (
    "github.com/Zentrix-Software-Hive/zyntax-ai-services/pkg/models"
    "gorm.io/gorm"
)
```
```shell
package models
import (
    "github.com/Zentrix-Software-Hive/zyntax-ai-services/pkg/user"
    "gorm.io/gorm"
)
```
### OR
```shell
    type User struct {
        ID       uint `json:"id" gorm:"primaryKey;autoIncrement"`
        FacultyID uint `json:"faculty_id" gorm:"not null"`
        Faculty   *models.Faculty `json:"faculty" gorm:"foreignKey:FacultyID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
    }
```
```shell
    type Faculty struct {
        ID       uint `json:"id" gorm:"primaryKey;autoIncrement"`
        UserID   uint `json:"user_id" gorm:"not null"`
        User     *models.User `json:"user" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
    }
```
## Do this
```shell
package user
import (
    "github.com/Zentrix-Software-Hive/zyntax-ai-services/pkg/models"
    "gorm.io/gorm"
)
```
```shell
package models
import (
    "gorm.io/gorm"
)
```
### OR
```shell
    type User struct {
        ID       uint `json:"id" gorm:"primaryKey;autoIncrement"`
    }
```
```shell
    type Faculty struct {
        ID       uint `json:"id" gorm:"primaryKey;autoIncrement"`
    }
```
```shell
    type MapUserFaculty struct {
        ID       uint `json:"id" gorm:"primaryKey;autoIncrement"`
        UserID   uint `json:"user_id" gorm:"not null"`
        User     *models.User `json:"user" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
        FacultyID uint `json:"faculty_id" gorm:"not null"`
        Faculty   *models.Faculty `json:"faculty" gorm:"foreignKey:FacultyID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
    }
```