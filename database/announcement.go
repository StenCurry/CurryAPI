package database

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrAnnouncementNotFound = errors.New("announcement not found")
)

// Announcement 公告模型
type Announcement struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedBy int64     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active"`
}

// AnnouncementRead 公告阅读记录模型
type AnnouncementRead struct {
	ID             int64     `json:"id"`
	AnnouncementID int64     `json:"announcement_id"`
	UserID         int64     `json:"user_id"`
	ReadAt         time.Time `json:"read_at"`
}

// AnnouncementWithReadStatus 带阅读状态的公告模型
type AnnouncementWithReadStatus struct {
	Announcement
	IsRead bool `json:"is_read"`
}

// CreateAnnouncement 创建新公告
func CreateAnnouncement(title, content string, createdBy int64) (*Announcement, error) {
	now := time.Now()
	result, err := db.Exec(
		`INSERT INTO announcements (title, content, created_by, created_at, updated_at, is_active) 
		 VALUES (?, ?, ?, ?, ?, ?)`,
		title, content, createdBy, now, now, true,
	)
	if err != nil {
		return nil, err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	
	return &Announcement{
		ID:        id,
		Title:     title,
		Content:   content,
		CreatedBy: createdBy,
		CreatedAt: now,
		UpdatedAt: now,
		IsActive:  true,
	}, nil
}

// GetAnnouncements 获取所有公告列表（按创建时间降序）
func GetAnnouncements(limit, offset int) ([]*Announcement, int, error) {
	// 获取总数
	var total int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM announcements WHERE is_active = TRUE`,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	// 获取公告列表
	rows, err := db.Query(
		`SELECT id, title, content, created_by, created_at, updated_at, is_active 
		 FROM announcements 
		 WHERE is_active = TRUE 
		 ORDER BY created_at DESC 
		 LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var announcements []*Announcement
	for rows.Next() {
		announcement := &Announcement{}
		err := rows.Scan(
			&announcement.ID,
			&announcement.Title,
			&announcement.Content,
			&announcement.CreatedBy,
			&announcement.CreatedAt,
			&announcement.UpdatedAt,
			&announcement.IsActive,
		)
		if err != nil {
			return nil, 0, err
		}
		announcements = append(announcements, announcement)
	}
	
	return announcements, total, nil
}

// GetAnnouncementByID 根据ID获取公告
func GetAnnouncementByID(id int64) (*Announcement, error) {
	announcement := &Announcement{}
	err := db.QueryRow(
		`SELECT id, title, content, created_by, created_at, updated_at, is_active 
		 FROM announcements WHERE id = ? AND is_active = TRUE`,
		id,
	).Scan(
		&announcement.ID,
		&announcement.Title,
		&announcement.Content,
		&announcement.CreatedBy,
		&announcement.CreatedAt,
		&announcement.UpdatedAt,
		&announcement.IsActive,
	)
	
	if err == sql.ErrNoRows {
		return nil, ErrAnnouncementNotFound
	}
	if err != nil {
		return nil, err
	}
	
	return announcement, nil
}

// DeleteAnnouncement 删除公告（软删除）
func DeleteAnnouncement(id int64) error {
	result, err := db.Exec(
		`UPDATE announcements SET is_active = FALSE WHERE id = ?`,
		id,
	)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return ErrAnnouncementNotFound
	}
	
	return nil
}

// MarkAsRead 标记公告为已读
func MarkAsRead(announcementID, userID int64) error {
	// 使用 INSERT IGNORE 实现幂等性
	_, err := db.Exec(
		`INSERT IGNORE INTO announcement_reads (announcement_id, user_id, read_at) 
		 VALUES (?, ?, ?)`,
		announcementID, userID, time.Now(),
	)
	return err
}

// GetUnreadCount 获取用户的未读公告数量
func GetUnreadCount(userID int64) (int, error) {
	var count int
	err := db.QueryRow(
		`SELECT COUNT(*) 
		 FROM announcements a 
		 WHERE a.is_active = TRUE 
		 AND NOT EXISTS (
			 SELECT 1 FROM announcement_reads ar 
			 WHERE ar.announcement_id = a.id AND ar.user_id = ?
		 )`,
		userID,
	).Scan(&count)
	
	if err != nil {
		return 0, err
	}
	
	return count, nil
}

// GetAnnouncementsWithReadStatus 获取带阅读状态的公告列表
func GetAnnouncementsWithReadStatus(userID int64, limit, offset int) ([]*AnnouncementWithReadStatus, int, error) {
	// 获取总数
	var total int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM announcements WHERE is_active = TRUE`,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	// 获取公告列表及阅读状态
	rows, err := db.Query(
		`SELECT 
			a.id, a.title, a.content, a.created_by, a.created_at, a.updated_at, a.is_active,
			CASE WHEN ar.id IS NOT NULL THEN TRUE ELSE FALSE END as is_read
		 FROM announcements a
		 LEFT JOIN announcement_reads ar ON a.id = ar.announcement_id AND ar.user_id = ?
		 WHERE a.is_active = TRUE
		 ORDER BY a.created_at DESC
		 LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var announcements []*AnnouncementWithReadStatus
	for rows.Next() {
		announcement := &AnnouncementWithReadStatus{}
		err := rows.Scan(
			&announcement.ID,
			&announcement.Title,
			&announcement.Content,
			&announcement.CreatedBy,
			&announcement.CreatedAt,
			&announcement.UpdatedAt,
			&announcement.IsActive,
			&announcement.IsRead,
		)
		if err != nil {
			return nil, 0, err
		}
		announcements = append(announcements, announcement)
	}
	
	return announcements, total, nil
}
