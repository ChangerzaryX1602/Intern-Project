package models

type Search struct {
	Keyword string `query:"keyword"`
	Column  string `query:"column"`
}

func (s *Search) GetSearchString() string {
	return "keyword=" + s.Keyword + "&column=" + s.Column
}

// AdmissionRoundFilter เพิ่ม filters สำหรับ admission rounds
type AdmissionRoundFilter struct {
	Search
	AcademicYear string `query:"academic_year"`
	Type         string `query:"type"`
}

// ProgramFilter เพิ่ม filters สำหรับ programs
type ProgramFilter struct {
	Search
	DepartmentID     uint `query:"department_id"`
	LevelID          uint `query:"level_id"`
	FacultyID        uint `query:"faculty_id"`
	AdmissionRoundID uint `query:"admission_round_id"`
}
