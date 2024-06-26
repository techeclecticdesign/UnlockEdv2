package database

import (
	"UnlockEdv2/src/models"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
)

func (db *DB) CreateActivity(activity *models.Activity) error {
	// This function is calling a stored procedure in the database to calculate the delta (src/activity_proc.sql)
	err := db.Conn.Exec("SELECT insert_daily_activity(?, ?, ?, ?, ?)",
		activity.UserID, activity.ProgramID, activity.Type, activity.TotalTime, activity.ExternalID).Error
	return LogDbError(err, "Failed to create activity.")
}

func (db *DB) GetActivityByUserID(page, perPage, userID int) (int64, []models.Activity, error) {
	var activities []models.Activity
	var count int64
	_ = db.Conn.Model(&models.Activity{}).Where("user_id = ?", userID).Count(&count)
	err := db.Conn.Where("user_id = ?", userID).Offset((page - 1) * perPage).Limit(perPage).Find(&activities).Error
	return count, activities, LogDbError(err, "Failed to get activities by user ID.")
}

func (db *DB) GetDailyActivityByUserID(userID int, year int) ([]models.DailyActivity, error) {
	var activities []models.Activity
	var startDate time.Time
	var endDate time.Time

	// Calculate start and end dates for the past year
	if year == 0 {
		startDate = time.Now().AddDate(-1, 0, 0).Truncate(24 * time.Hour)
		endDate = time.Now().Truncate(24 * time.Hour)
	} else {
		startDate = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate = time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC).Add(-1 * time.Second)
	}

	if err := db.Conn.Where("user_id = ? AND created_at BETWEEN ? AND ?", userID, startDate, endDate).Find(&activities).Error; err != nil {
		return nil, LogDbError(err, "Failed to get activity for date range.")
	}

	// Combine activities based on date
	dailyActivities := make(map[time.Time]models.DailyActivity)
	for _, activity := range activities {
		date := activity.CreatedAt.Truncate(24 * time.Hour)
		if dailyActivity, ok := dailyActivities[date]; ok {
			dailyActivity.TotalTime += activity.TimeDelta
			dailyActivity.Activities = append(dailyActivity.Activities, activity)
			dailyActivities[date] = dailyActivity
		} else {
			dailyActivities[date] = models.DailyActivity{
				Date:       date,
				TotalTime:  activity.TimeDelta,
				Activities: []models.Activity{activity},
			}
		}
	}

	// Convert map to slice
	var dailyActivityList []models.DailyActivity
	for _, dailyActivity := range dailyActivities {
		dailyActivityList = append(dailyActivityList, dailyActivity)
	}

	// Sort daily activity list by total time
	sort.Slice(dailyActivityList, func(i, j int) bool {
		return dailyActivityList[i].TotalTime < dailyActivityList[j].TotalTime
	})

	// Calculate quartiles for each day's activities
	n := len(dailyActivityList)
	for i := range dailyActivityList {
		switch {
		case i < n/4:
			dailyActivityList[i].Quartile = 1
		case i < n/2:
			dailyActivityList[i].Quartile = 2
		case i < 3*n/4:
			dailyActivityList[i].Quartile = 3
		default:
			dailyActivityList[i].Quartile = 4
		}
	}

	//sort by date
	sort.Slice(dailyActivityList, func(i, j int) bool {
		return dailyActivityList[i].Date.Before(dailyActivityList[j].Date)
	})

	return dailyActivityList, nil
}

func (db *DB) GetActivityByProgramID(page, perPage, programID int) (int64, []models.Activity, error) {
	var activities []models.Activity
	var count int64
	_ = db.Conn.Model(&models.Activity{}).Where("program_id = ?", programID).Count(&count)
	err := db.Conn.Where("program_id = ?", programID).Offset((page - 1) * perPage).Limit(perPage).Find(&activities).Error
	return count, activities, LogDbError(err, "Failed to get activities by program ID.")
}

func (db *DB) DeleteActivity(activityID int) error {
	err := db.Conn.Delete(&models.Activity{}, activityID).Error
	return LogDbError(err, "Failed to delete activity.")
}

type result struct {
	ProgramID            uint
	AltName              string
	Name                 string
	ProviderPlatformName string
	ExternalURL          string
	Date                 string
	TimeDelta            uint
}

func dashboardHelper(results []result) []models.CurrentEnrollment {
	enrollmentsMap := make(map[uint]*models.CurrentEnrollment)
	for _, result := range results {
		if _, exists := enrollmentsMap[result.ProgramID]; !exists {
			enrollmentsMap[result.ProgramID] = &models.CurrentEnrollment{
				AltName:              result.AltName,
				Name:                 result.Name,
				ProviderPlatformName: result.ProviderPlatformName,
				ExternalURL:          result.ExternalURL,
				TotalTime:            0,
			}
		}
		enrollmentsMap[result.ProgramID].TotalTime += result.TimeDelta
	}

	var enrollments []models.CurrentEnrollment
	for _, enrollment := range enrollmentsMap {
		enrollments = append(enrollments, *enrollment)
	}
	return enrollments
}

func (db *DB) GetUserDashboardInfo(userID int) (models.UserDashboardJoin, error) {
	var recentPrograms [3]models.RecentProgram
	// first get the users 3 most recently interacted with programs
	//
	// get the users 3 most recent programs. progress: # of milestones where type = "assignment_submission" && status = "complete" || "graded" +
	// # of milestones where type = "quiz_submission" && status = "complete" || "graded"
	// where the user has an entry in the activity table
	err := db.Conn.Table("programs p").
		Select(`p.name as program_name,
        p.alt_name,
        p.thumbnail_url,
        p.external_url,
        pp.name as provider_platform_name,
        COUNT(sub_milestones.id) * 100.0 / p.total_progress_milestones as course_progress`).
		Joins(`LEFT JOIN (
    SELECT id, program_id FROM milestones
    WHERE user_id = ? AND type IN ('assignment_submission', 'quiz_submission') AND is_completed = true
    ORDER BY created_at DESC
    LIMIT 3 ) sub_milestones ON sub_milestones.program_id = p.id`, userID).
		Joins("LEFT JOIN provider_platforms pp ON p.provider_platform_id = pp.id").
		Joins("LEFT JOIN outcomes o ON o.program_id = p.id AND o.user_id = ?", userID).
		Where("p.id IN (SELECT program_id FROM activities WHERE user_id = ?)", userID).
		Where("o.type IS NULL").
		Group("p.id, p.name, p.alt_name, p.thumbnail_url, pp.name").
		Find(&recentPrograms).Error
	if err != nil {
		LogDbError(err, "Failed to get recent programs for user.")
		recentPrograms = [3]models.RecentProgram{}
	}
	// then get the users current enrollments
	// which is the program.alt_name, program.name, provider_platform.name as provider_platform_name (join on provider_platform_id = provider_platform.id), program.external_url,
	// and acitvity.total_activity_time for the last 7 days of activity where activity.user_id = userID and activity.program_id = program.id
	results := []result{}

	err = db.Conn.Table("programs p").
		Select(`p.id as program_id,
				p.alt_name,
				p.name,
				pp.name as provider_platform_name,
				p.external_url,
				DATE(a.created_at) as date,
				SUM(a.time_delta) as time_delta`).
		Joins("JOIN provider_platforms pp ON p.provider_platform_id = pp.id").
		Joins("JOIN activities a ON a.program_id = p.id").
		Joins("LEFT JOIN outcomes o ON o.program_id = p.id AND o.user_id = ?", userID).
		Where("a.user_id = ? AND a.created_at >= ?", userID, time.Now().AddDate(0, 0, -7)).
		Where("o.type IS NULL").
		Group("p.id, DATE(a.created_at), pp.name").
		Find(&results).Error
	if err != nil {
		log.Fatalf("Query failed when getting enrollments: %v", err)
	}
	enrollments := dashboardHelper(results)
	if len(enrollments) == 0 {
		log.Info("No enrollments found.")
		var newEnrollments []struct {
			ProgramID            uint
			AltName              string
			Name                 string
			ProviderPlatformName string
			ExternalURL          string
		}
		err = db.Conn.Table("programs p").Select(`p.id as program_id, p.alt_name, p.name, pp.name as provider_platform_name, p.external_url`).
			Joins(`JOIN provider_platforms pp ON p.provider_platform_id = pp.id`).
			Joins(`LEFT JOIN milestones m on m.program_id = p.id`).Where(`m.user_id`, userID).Find(&newEnrollments).Error
		if err != nil {
			log.Fatalf("Query failed when getting new enrollments:: %v", err)
		}
		for idx, enrollment := range newEnrollments {
			if idx == 7 {
				break
			}
			enrollments = append(enrollments, models.CurrentEnrollment{
				AltName:              enrollment.AltName,
				Name:                 enrollment.Name,
				ProviderPlatformName: enrollment.ProviderPlatformName,
				ExternalURL:          enrollment.ExternalURL,
			})
			log.Debugf("enrollments: %v", enrollments)
		}
	}
	// get activity for past 7 days
	var activities []models.RecentActivity

	err = db.Conn.Table("activities a").
		Select(`DATE(a.created_at) as date, SUM(a.time_delta) as delta`).
		Where("a.user_id = ? AND a.created_at >= ?", userID, time.Now().AddDate(0, 0, -7)).
		Group("date").
		Find(&activities).Error
	if err != nil {
		log.Fatalf("Query failed getting 7 days previous activity: %v", err)
	}

	return models.UserDashboardJoin{Enrollments: enrollments, RecentPrograms: recentPrograms, WeekActivity: activities}, err
}
