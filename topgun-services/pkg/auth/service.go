package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"topgun-services/pkg/domain"
	"topgun-services/pkg/models"
	"topgun-services/pkg/utils"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	helpers "github.com/zercle/gofiber-helpers"
	"gorm.io/gorm"
)

type authService struct {
	domain.AuthRepository
	domain.UserRepository
}

func NewAuthService(authRepo domain.AuthRepository, userRepo domain.UserRepository) domain.AuthService {
	return &authService{
		AuthRepository: authRepo,
		UserRepository: userRepo,
	}
}
func (s *authService) SignToken(user *models.User, host string) (string, error) {
	token, err := s.AuthRepository.SignToken(user, host)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *authService) ResetPassword(userID uuid.UUID, body *models.ResetPasswordRequest) []helpers.ResponseError {
	user, err := s.UserRepository.GetUser(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []helpers.ResponseError{
				helpers.ResponseError(*helpers.NewError(http.StatusNotFound, "User not found")),
			}
		}
		return []helpers.ResponseError{
			helpers.ResponseError(*helpers.NewError(http.StatusInternalServerError, "Failed to retrieve user: "+err.Error())),
		}
	}

	if user.Password == "" {
		return []helpers.ResponseError{
			helpers.ResponseError(*helpers.NewError(http.StatusBadRequest, "Password reset not available for OAuth users")),
		}
	}

	if !user.CheckPassword(body.OldPassword) {
		return []helpers.ResponseError{
			helpers.ResponseError(*helpers.NewError(http.StatusBadRequest, "Old password is incorrect")),
		}
	}

	err = s.UserRepository.UpdateUserPassword(userID, body.NewPassword)
	if err != nil {
		return []helpers.ResponseError{
			helpers.ResponseError(*helpers.NewError(http.StatusInternalServerError, "Failed to update password: "+err.Error())),
		}
	}

	return nil
}
func (s *authService) VerifyTokenAndResetPassword(token string, password string) error {
	userId, err := s.AuthRepository.VerifyToken(token)
	if err != nil {
		return err
	}
	err = s.UserRepository.UpdateUserPassword(userId, password)
	if err != nil {
		return err
	}
	return nil
}
func (s *authService) IssueTokenVerification(email string, host string) (*string, error) {
	emailString := string(email)
	emailString = strings.ReplaceAll(emailString, "%40", "@")
	token, err := s.AuthRepository.IssueTokenVerification(emailString)
	if err != nil {
		return nil, err
	}
	if viper.GetString("app.env") == "development" {
		host = "http://localhost:4200"
	}
	url := fmt.Sprintf("%s/user/recovery?token=%s", host, *token)
	sendMail := utils.SendMailModel{
		SenderName:   "GS-Admission",
		SenderMail:   viper.GetString("app.smtp.username"),
		ReceiverName: emailString,
		ReceiverMail: emailString,
		Subject:      "Reset your password for GS-Admission",
		Body: fmt.Sprintf(`
<html>
  <body style="font-family: Tahoma, Arial, sans-serif; background-color: #f9f9f9; padding: 20px;">
    <table width="100%%" cellpadding="0" cellspacing="0">
      <tr>
        <td align="center">
          <table width="600" cellpadding="0" cellspacing="0" style="background-color: #ffffff; padding: 30px; border-radius: 8px; box-shadow: 0 2px 6px rgba(0,0,0,0.1);">
            
            <!-- โลโก้ -->
            <tr>
              <td align="center" style="padding-bottom: 20px;">
                <img src="https://app.gs.kku.ac.th/gs/uikit/img/newLogo.png" alt="โลโก้" width="120" style="display:block;">
              </td>
            </tr>
            
            <!-- หัวข้อ -->
            <tr>
              <td align="center" style="padding-bottom: 20px;">
                <h2 style="color: #ff6600; margin: 0;">ลืมรหัสผ่านใช่หรือไม่?</h2>
              </td>
            </tr>
            
            <!-- ข้อความ -->
            <tr>
              <td style="color: #555; font-size: 15px; line-height: 22px; padding-bottom: 25px;">
                <p>เราได้รับคำขอให้รีเซ็ตรหัสผ่านของคุณ กรุณาคลิกปุ่มด้านล่างเพื่อสร้างรหัสผ่านใหม่:</p>
              </td>
            </tr>
            
            <!-- ปุ่ม -->
            <tr>
              <td align="center" style="padding-bottom: 30px;">
                <a href="%s" style="background-color: #ff6600; color: #ffffff; text-decoration: none; padding: 12px 24px; border-radius: 6px; font-size: 16px; display: inline-block;">
                  รีเซ็ตรหัสผ่าน
                </a>
              </td>
            </tr>
            
            <!-- ข้อความเพิ่มเติม -->
            <tr>
              <td style="color: #999; font-size: 13px; line-height: 20px; text-align: center;">
                <p>หากคุณไม่ได้ส่งคำขอรีเซ็ตรหัสผ่าน กรุณาละเว้นอีเมลฉบับนี้ได้เลย</p>
              </td>
            </tr>
          
          </table>
        </td>
      </tr>
    </table>
  </body>
</html>
`, url),
	}
	if viper.GetString("app.env") == "development" {
		_, err = utils.SendNormalMail(sendMail)
		if err != nil {
			return nil, err
		}
	} else {
		_, err = utils.SendMail(sendMail)
		if err != nil {
			return nil, err
		}
	}
	return &url, nil
}
