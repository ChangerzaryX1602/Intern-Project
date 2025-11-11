package infrastructure

import "topgun-services/pkg/models"

func (s *Server) AutoMigrate() (err error) {
	if err = s.MainDbConn.AutoMigrate(
		models.User{},
		models.Camera{},
		models.Detect{},
	); err != nil {
		return
	}
	if err = s.LogDbConn.AutoMigrate(
		models.Log{},
	); err != nil {
		return
	}
	return
}
