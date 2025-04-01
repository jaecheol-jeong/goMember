package db

import (
	"database/sql"
	"log"
	"membership/internal/models"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MemberDB는 members 테이블에 대한 데이터베이스 작업을 처리
type MemberDB struct {
	Db *sql.DB
}

// NewMemberDB는 새로운 MemberDB 인스턴스를 생성합니다.
func NewMemberDB(dataSourceName string) (*MemberDB, error) {
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Println(dataSourceName)
		return nil, err
	}
	// 연결 테스트
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &MemberDB{Db: db}, nil
}

// InsertMember는 새로운 멤버를 추가
func (m *MemberDB) InsertMember(member models.Member) error {
	err := member.HashPassword(member.Password)
	if err != nil {
		return err
	}
	query := "INSERT INTO members (id, email, name, password, role, created_at) VALUES (?, ?, ?, ?, ?, ?)"
	_, err = m.Db.Exec(query, member.ID, member.Email, member.Name, member.Password, member.Role, member.CreatedAt)
	return err
}

// UpdateMember는 기존 멤버 정보를 수정
func (m *MemberDB) UpdateMember(member models.Member) error {
	query := "UPDATE members SET email = ?, name = ?, password = ?, role = ?, created_at = ? WHERE id = ?"
	_, err := m.Db.Exec(query, member.Email, member.Name, member.Password, member.Role, member.CreatedAt, member.ID)
	return err
}

// DeleteMember는 멤버를 삭제
func (m *MemberDB) DeleteMember(id string) error {
	query := "DELETE FROM members WHERE id = ?"
	_, err := m.Db.Exec(query, id)
	return err
}

// GetMember는 멤버 ID로 멤버 정보를 조회
func (m *MemberDB) GetMember(id string) (*models.Member, error) {
	query := "SELECT id, email, name, password, role, created_at FROM members WHERE id = ?"
	row := m.Db.QueryRow(query, id)

	var member models.Member
	var createdAt []uint8 // 임시로 []uint8 타입으로 받음
	err := row.Scan(&member.ID, &member.Email, &member.Name, &member.Password, &member.Role, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 멤버가 없을 경우 nil 반환
		}
		return nil, err
	}

	// []uint8 타입의 createdAt을 time.Time으로 변환
	member.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return nil, err
	}

	return &member, nil
}

// SearchMembers는 검색 조건에 맞는 멤버 목록을 조회
func (m *MemberDB) SearchMembers(name string) ([]*models.Member, error) {
	query := "SELECT id, email, name, password, role, created_at FROM members WHERE name LIKE ?"
	rows, err := m.Db.Query(query, "%"+name+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.Member
	for rows.Next() {
		var member models.Member
		var createdAt []uint8 // 임시로 []uint8 타입으로 받음
		err := rows.Scan(&member.ID, &member.Email, &member.Name, &member.Password, &member.Role, &createdAt)
		if err != nil {
			return nil, err
		}

		// []uint8 타입의 createdAt을 time.Time으로 변환
		member.CreatedAt, err = parseTime(createdAt)
		if err != nil {
			return nil, err
		}
		members = append(members, &member)
	}

	return members, nil
}

// parseTime은 []uint8 타입의 시간을 time.Time 타입으로 변환합니다.
func parseTime(t []uint8) (time.Time, error) {
	timeString := string(t)
	return time.Parse("2006-01-02 15:04:05", timeString) // MySQL datetime format
}

// Close는 데이터베이스
func (m *MemberDB) Close() error {
	return m.Db.Close()
}
