package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"home-decision/backend/internal/model"
)

var ErrHouseNotFound = errors.New("house not found")
var ErrUserConflict = errors.New("user conflict")

type MySQLStore struct {
	db   *sql.DB
	meta model.Meta
}

func NewMySQLStore(dsn string, meta model.Meta) (*MySQLStore, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	return &MySQLStore{db: db, meta: meta}, nil
}

func (s *MySQLStore) Close() error {
	return s.db.Close()
}

func (s *MySQLStore) GetMeta() model.Meta {
	return s.meta
}

func (s *MySQLStore) GetWeights(householdID string) ([]model.WeightProfile, error) {
	if err := s.ensureHousehold(householdID); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(`
		SELECT role_code, metric_key, weight_value
		FROM weight_profiles
		WHERE household_id = ?
		ORDER BY role_code, metric_key
	`, householdID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profiles := map[string]*model.WeightProfile{}
	for rows.Next() {
		var role, metric string
		var value float64
		if err := rows.Scan(&role, &metric, &value); err != nil {
			return nil, err
		}
		if _, ok := profiles[role]; !ok {
			profiles[role] = &model.WeightProfile{
				Role:    role,
				Label:   defaultProfileLabel(role),
				Weights: map[string]float64{},
			}
		}
		profiles[role].Weights[metric] = value
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(profiles) == 0 {
		return defaultProfiles(), nil
	}

	items := make([]model.WeightProfile, 0, len(profiles))
	for _, role := range []string{"me", "partner"} {
		if profile, ok := profiles[role]; ok {
			items = append(items, *profile)
		}
	}
	for role, profile := range profiles {
		if role != "me" && role != "partner" {
			items = append(items, *profile)
		}
	}
	return items, nil
}

func (s *MySQLStore) SaveWeights(householdID string, profiles []model.WeightProfile) error {
	if err := s.ensureHousehold(householdID); err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM weight_profiles WHERE household_id = ?`, householdID); err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO weight_profiles (household_id, role_code, metric_key, weight_value)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, profile := range profiles {
		for metric, value := range profile.Weights {
			if _, err := stmt.Exec(householdID, profile.Role, metric, value); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (s *MySQLStore) ListHouses(householdID string) ([]model.House, error) {
	if err := s.ensureHousehold(householdID); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(`
		SELECT id, household_id, community_name, listing_name, view_date,
		       total_price, unit_price, area, house_age, floor_text, orientation,
		       house_type, renovation, commute_time, metro_time, monthly_fee,
		       living_convenience, efficiency_rate, light_score, noise_score,
		       layout_score, property_score, community_score, comfort_score,
		       parking_score, notes
		FROM houses
		WHERE household_id = ?
		ORDER BY updated_at DESC, created_at DESC
	`, householdID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var houses []model.House
	for rows.Next() {
		house, err := scanHouse(rows)
		if err != nil {
			return nil, err
		}
		houses = append(houses, house)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := s.attachSelections(houses); err != nil {
		return nil, err
	}
	return houses, nil
}

func (s *MySQLStore) GetHouse(householdID, houseID string) (*model.House, error) {
	if err := s.ensureHousehold(householdID); err != nil {
		return nil, err
	}

	row := s.db.QueryRow(`
		SELECT id, household_id, community_name, listing_name, view_date,
		       total_price, unit_price, area, house_age, floor_text, orientation,
		       house_type, renovation, commute_time, metro_time, monthly_fee,
		       living_convenience, efficiency_rate, light_score, noise_score,
		       layout_score, property_score, community_score, comfort_score,
		       parking_score, notes
		FROM houses
		WHERE household_id = ? AND id = ?
	`, householdID, houseID)

	house, err := scanHouse(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrHouseNotFound
		}
		return nil, err
	}

	houses := []model.House{house}
	if err := s.attachSelections(houses); err != nil {
		return nil, err
	}
	return &houses[0], nil
}

func (s *MySQLStore) CreateHouse(householdID string, house model.House) (*model.House, error) {
	if err := s.ensureHousehold(householdID); err != nil {
		return nil, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := upsertHouse(tx, householdID, house, true); err != nil {
		return nil, err
	}
	if err := replaceSelections(tx, house.ID, house.BonusSelections, "house_bonus_selections", "bonus_key"); err != nil {
		return nil, err
	}
	if err := replaceSelections(tx, house.ID, house.RiskSelections, "house_risk_selections", "risk_key"); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.GetHouse(householdID, house.ID)
}

func (s *MySQLStore) UpdateHouse(householdID, houseID string, house model.House) (*model.House, error) {
	if err := s.ensureHousehold(householdID); err != nil {
		return nil, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	house.ID = houseID
	if err := upsertHouse(tx, householdID, house, false); err != nil {
		return nil, err
	}
	if err := replaceSelections(tx, houseID, house.BonusSelections, "house_bonus_selections", "bonus_key"); err != nil {
		return nil, err
	}
	if err := replaceSelections(tx, houseID, house.RiskSelections, "house_risk_selections", "risk_key"); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.GetHouse(householdID, houseID)
}

func (s *MySQLStore) DeleteHouse(householdID, houseID string) error {
	if err := s.ensureHousehold(householdID); err != nil {
		return err
	}

	result, err := s.db.Exec(`DELETE FROM houses WHERE household_id = ? AND id = ?`, householdID, houseID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrHouseNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanHouse(s scanner) (model.House, error) {
	var house model.House
	var viewDate sql.NullTime
	var notes sql.NullString
	err := s.Scan(
		&house.ID,
		&house.HouseholdID,
		&house.CommunityName,
		&house.ListingName,
		&viewDate,
		&house.TotalPrice,
		&house.UnitPrice,
		&house.Area,
		&house.HouseAge,
		&house.Floor,
		&house.Orientation,
		&house.HouseType,
		&house.Renovation,
		&house.CommuteTime,
		&house.MetroTime,
		&house.MonthlyFee,
		&house.LivingConvenience,
		&house.EfficiencyRate,
		&house.LightScore,
		&house.NoiseScore,
		&house.LayoutScore,
		&house.PropertyScore,
		&house.CommunityScore,
		&house.ComfortScore,
		&house.ParkingScore,
		&notes,
	)
	if err != nil {
		return model.House{}, err
	}
	if viewDate.Valid {
		house.ViewDate = viewDate.Time.Format("2006-01-02")
	}
	if notes.Valid {
		house.Notes = notes.String
	}
	return house, nil
}

func (s *MySQLStore) attachSelections(houses []model.House) error {
	for index := range houses {
		bonus, err := s.fetchSelections(houses[index].ID, "house_bonus_selections", "bonus_key")
		if err != nil {
			return err
		}
		risk, err := s.fetchSelections(houses[index].ID, "house_risk_selections", "risk_key")
		if err != nil {
			return err
		}
		houses[index].BonusSelections = bonus
		houses[index].RiskSelections = risk
	}
	return nil
}

func (s *MySQLStore) fetchSelections(houseID, table, column string) ([]string, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE house_id = ? ORDER BY id ASC`, column, table)
	rows, err := s.db.Query(query, houseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		items = append(items, value)
	}
	return items, rows.Err()
}

func upsertHouse(tx *sql.Tx, householdID string, house model.House, isCreate bool) error {
	if isCreate {
		_, err := tx.Exec(`
			INSERT INTO houses (
				id, household_id, community_name, listing_name, view_date,
				total_price, unit_price, area, house_age, floor_text, orientation,
				house_type, renovation, commute_time, metro_time, monthly_fee,
				living_convenience, efficiency_rate, light_score, noise_score,
				layout_score, property_score, community_score, comfort_score,
				parking_score, notes
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			house.ID, householdID, house.CommunityName, house.ListingName, nullableDate(house.ViewDate),
			house.TotalPrice, house.UnitPrice, house.Area, house.HouseAge, house.Floor, house.Orientation,
			house.HouseType, house.Renovation, house.CommuteTime, house.MetroTime, house.MonthlyFee,
			house.LivingConvenience, house.EfficiencyRate, house.LightScore, house.NoiseScore,
			house.LayoutScore, house.PropertyScore, house.CommunityScore, house.ComfortScore,
			house.ParkingScore, nullableString(house.Notes),
		)
		return err
	}

	result, err := tx.Exec(`
		UPDATE houses
		SET community_name = ?, listing_name = ?, view_date = ?, total_price = ?, unit_price = ?,
		    area = ?, house_age = ?, floor_text = ?, orientation = ?, house_type = ?, renovation = ?,
		    commute_time = ?, metro_time = ?, monthly_fee = ?, living_convenience = ?, efficiency_rate = ?,
		    light_score = ?, noise_score = ?, layout_score = ?, property_score = ?, community_score = ?,
		    comfort_score = ?, parking_score = ?, notes = ?
		WHERE household_id = ? AND id = ?
	`,
		house.CommunityName, house.ListingName, nullableDate(house.ViewDate), house.TotalPrice, house.UnitPrice,
		house.Area, house.HouseAge, house.Floor, house.Orientation, house.HouseType, house.Renovation,
		house.CommuteTime, house.MetroTime, house.MonthlyFee, house.LivingConvenience, house.EfficiencyRate,
		house.LightScore, house.NoiseScore, house.LayoutScore, house.PropertyScore, house.CommunityScore,
		house.ComfortScore, house.ParkingScore, nullableString(house.Notes), householdID, house.ID,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrHouseNotFound
	}
	return nil
}

func replaceSelections(tx *sql.Tx, houseID string, selections []string, table, column string) error {
	deleteQuery := fmt.Sprintf(`DELETE FROM %s WHERE house_id = ?`, table)
	if _, err := tx.Exec(deleteQuery, houseID); err != nil {
		return err
	}
	if len(selections) == 0 {
		return nil
	}

	insertQuery := fmt.Sprintf(`INSERT INTO %s (house_id, %s) VALUES (?, ?)`, table, column)
	stmt, err := tx.Prepare(insertQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, value := range selections {
		if _, err := stmt.Exec(houseID, value); err != nil {
			return err
		}
	}
	return nil
}

func nullableDate(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func defaultProfileLabel(role string) string {
	switch role {
	case "me":
		return "我的偏好"
	case "partner":
		return "另一半偏好"
	default:
		return role
	}
}

func (s *MySQLStore) ensureHousehold(householdID string) error {
	if _, err := s.db.Exec(`
		INSERT INTO households (id, name)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE name = VALUES(name)
	`, householdID, "默认家庭"); err != nil {
		return err
	}

	for _, member := range []struct {
		role  string
		label string
	}{
		{role: "me", label: "我"},
		{role: "partner", label: "另一半"},
	} {
		if _, err := s.db.Exec(`
			INSERT INTO household_members (household_id, role_code, display_name)
			VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE display_name = VALUES(display_name)
		`, householdID, member.role, member.label); err != nil {
			return err
		}
	}
	return nil
}

func defaultProfiles() []model.WeightProfile {
	return []model.WeightProfile{
		{
			Role:  "me",
			Label: "我的偏好",
			Weights: map[string]float64{
				"totalPrice": 25, "commuteTime": 22, "houseAge": 12, "houseTypeScore": 8,
				"layoutScore": 10, "lightScore": 8, "noiseScore": 5, "communityScore": 5,
				"renovationScore": 2, "propertyScore": 1, "parkingScore": 1, "livingConvenience": 1,
				"comfortScore": 6, "efficiencyRate": 1,
			},
		},
		{
			Role:  "partner",
			Label: "另一半偏好",
			Weights: map[string]float64{
				"totalPrice": 10, "commuteTime": 10, "houseAge": 5, "houseTypeScore": 5,
				"layoutScore": 20, "lightScore": 20, "noiseScore": 8, "communityScore": 10,
				"renovationScore": 7, "propertyScore": 2, "parkingScore": 1, "livingConvenience": 3,
				"comfortScore": 15, "efficiencyRate": 2,
			},
		},
	}
}

func (s *MySQLStore) CreateUser(user model.User, passwordSalt, passwordHash, householdID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var adminCount int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM users WHERE is_admin = 1`).Scan(&adminCount); err != nil {
		return err
	}
	if adminCount == 0 {
		user.IsAdmin = true
	}

	if _, err := tx.Exec(`INSERT INTO households (id, name) VALUES (?, ?)`, householdID, user.DisplayName+" 的家庭"); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		INSERT INTO users (id, login_id, display_name, link_code, is_admin, password_salt, password_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, user.ID, user.LoginID, user.DisplayName, user.LinkCode, user.IsAdmin, passwordSalt, passwordHash); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		INSERT INTO user_household_links (user_id, household_id)
		VALUES (?, ?)
	`, user.ID, householdID); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		INSERT INTO household_members (household_id, role_code, display_name)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE display_name = VALUES(display_name)
	`, householdID, "me", user.DisplayName); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *MySQLStore) FindUserAuthByLoginID(loginID string) (*UserAuthRecord, error) {
	row := s.db.QueryRow(`
		SELECT id, login_id, display_name, link_code, is_admin, password_salt, password_hash
		FROM users
		WHERE login_id = ?
	`, loginID)
	record, err := scanUserAuth(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (s *MySQLStore) FindUserByLoginID(loginID string) (*model.User, error) {
	record, err := s.FindUserAuthByLoginID(loginID)
	if err != nil || record == nil {
		return nil, err
	}
	user := record.User
	return &user, nil
}

func (s *MySQLStore) FindUserByID(userID string) (*model.User, error) {
	row := s.db.QueryRow(`
		SELECT id, login_id, display_name, link_code, is_admin
		FROM users
		WHERE id = ?
	`, userID)
	user, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *MySQLStore) FindUserByLinkCode(linkCode string) (*model.User, error) {
	row := s.db.QueryRow(`
		SELECT id, login_id, display_name, link_code, is_admin
		FROM users
		WHERE link_code = ?
	`, linkCode)
	user, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *MySQLStore) CreateSession(session model.Session) error {
	_, err := s.db.Exec(`
		INSERT INTO auth_sessions (token, user_id, expires_at)
		VALUES (?, ?, ?)
	`, session.Token, session.UserID, session.ExpiresAt)
	return err
}

func (s *MySQLStore) FindUserBySessionToken(token string) (*model.User, error) {
	row := s.db.QueryRow(`
		SELECT u.id, u.login_id, u.display_name, u.link_code, u.is_admin
		FROM auth_sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token = ? AND s.expires_at > NOW()
	`, token)
	user, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *MySQLStore) DeleteSession(token string) error {
	_, err := s.db.Exec(`DELETE FROM auth_sessions WHERE token = ?`, token)
	return err
}

func (s *MySQLStore) FindHouseholdIDByUserID(userID string) (string, error) {
	row := s.db.QueryRow(`SELECT household_id FROM user_household_links WHERE user_id = ?`, userID)
	var householdID string
	if err := row.Scan(&householdID); err != nil {
		return "", err
	}
	return householdID, nil
}

func (s *MySQLStore) ListAccountMembers(householdID string) ([]model.AccountMember, error) {
	rows, err := s.db.Query(`
		SELECT u.id, u.login_id, u.display_name, u.link_code
		FROM user_household_links uhl
		JOIN users u ON u.id = uhl.user_id
		WHERE uhl.household_id = ?
		ORDER BY u.created_at ASC
	`, householdID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.AccountMember
	for rows.Next() {
		var item model.AccountMember
		if err := rows.Scan(&item.UserID, &item.LoginID, &item.DisplayName, &item.LinkCode); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *MySQLStore) LinkUsers(userID, partnerUserID string) error {
	myHouseholdID, err := s.FindHouseholdIDByUserID(userID)
	if err != nil {
		return err
	}
	partnerHouseholdID, err := s.FindHouseholdIDByUserID(partnerUserID)
	if err != nil {
		return err
	}
	if myHouseholdID == partnerHouseholdID {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		UPDATE user_household_links
		SET household_id = ?
		WHERE user_id = ?
	`, myHouseholdID, partnerUserID); err != nil {
		return err
	}

	partnerUser, err := s.FindUserByID(partnerUserID)
	if err != nil {
		return err
	}
	if partnerUser != nil {
		if _, err := tx.Exec(`
			INSERT INTO household_members (household_id, role_code, display_name)
			VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE display_name = VALUES(display_name)
		`, myHouseholdID, "partner", partnerUser.DisplayName); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *MySQLStore) ListAdminUsers() ([]model.AdminUser, error) {
	rows, err := s.db.Query(`
		SELECT u.id, u.login_id, u.display_name, u.link_code, u.is_admin, uhl.household_id, u.created_at
		FROM users u
		LEFT JOIN user_household_links uhl ON uhl.user_id = u.id
		ORDER BY u.created_at ASC, u.id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.AdminUser
	for rows.Next() {
		var item model.AdminUser
		if err := rows.Scan(
			&item.ID,
			&item.LoginID,
			&item.DisplayName,
			&item.LinkCode,
			&item.IsAdmin,
			&item.HouseholdID,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *MySQLStore) SetUserAdmin(userID string, isAdmin bool) error {
	_, err := s.db.Exec(`UPDATE users SET is_admin = ? WHERE id = ?`, isAdmin, userID)
	return err
}

func scanUser(s scanner) (model.User, error) {
	var user model.User
	err := s.Scan(&user.ID, &user.LoginID, &user.DisplayName, &user.LinkCode, &user.IsAdmin)
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

func scanUserAuth(s scanner) (UserAuthRecord, error) {
	var record UserAuthRecord
	err := s.Scan(
		&record.User.ID,
		&record.User.LoginID,
		&record.User.DisplayName,
		&record.User.LinkCode,
		&record.User.IsAdmin,
		&record.PasswordSalt,
		&record.PasswordHash,
	)
	if err != nil {
		return UserAuthRecord{}, err
	}
	return record, nil
}
