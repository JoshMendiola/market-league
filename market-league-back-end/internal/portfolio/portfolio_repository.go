package portfolio

import (
	"fmt"
	"time"

	"github.com/market-league/internal/models"
	"gorm.io/gorm"
)

// PortfolioRepository provides access to portfolio-related operations in the database.
type PortfolioRepository struct {
	db *gorm.DB
}

// NewPortfolioRepository creates a new instance of PortfolioRepository.
func NewPortfolioRepository(db *gorm.DB) *PortfolioRepository {
	return &PortfolioRepository{db: db}
}

// GetPortfolioByID fetches a portfolio by its ID.
func (r *PortfolioRepository) GetPortfolioWithID(portfolioID uint) (*models.Portfolio, error) {
	var portfolio models.Portfolio
	// Preload the User, League, and Stocks associations
	err := r.db.
		Preload("User").
		Preload("League").
		Preload("Stocks").
		First(&portfolio, portfolioID).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("portfolio with ID %d not found", portfolioID)
		}
		return nil, fmt.Errorf("failed to fetch portfolio: %w", err)
	}

	return &portfolio, nil
}

// UpdatePortfolioPoints update points of a portfolio specifically
func (r *PortfolioRepository) GetAllPortfolios() ([]models.Portfolio, error) {
	var portfolios []models.Portfolio
	err := r.db.
		Preload("User").
		Preload("League").
		Preload("Stocks").
		Find(&portfolios).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve portfolios: %v", err)
	}
	return portfolios, nil
}

// GetPortfoliosForLeague gets all the portfolios and filters by a league
func (r *PortfolioRepository) GetPortfoliosForLeague(leagueID uint) ([]models.Portfolio, error) {
	var portfolios []models.Portfolio
	err := r.db.
		Preload("Stocks").
		Where("league_id = ?", leagueID).
		Find(&portfolios).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve portfolios: %v", err)
	}
	return portfolios, nil
}

// GetPortfolioIDByUserAndLeague retrieves the portfolio ID for a given user and league.
func (r *PortfolioRepository) GetPortfolioIDByUserAndLeague(userID, leagueID uint) (uint, error) {
	var portfolio models.Portfolio
	err := r.db.Select("id").Where("user_id = ? AND league_id = ?", userID, leagueID).First(&portfolio).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, fmt.Errorf("portfolio not found for user ID %d in league ID %d", userID, leagueID)
		}
		return 0, fmt.Errorf("failed to fetch portfolio ID: %w", err)
	}
	return portfolio.ID, nil
}

// CreatePortfolio creates a new portfolio for a user in a league.
func (r *PortfolioRepository) CreatePortfolio(portfolio *models.Portfolio) error {
	return r.db.Create(portfolio).Error
}

// UpdatePortfolio updates an existing portfolio in the database.
func (r *PortfolioRepository) UpdatePortfolio(portfolio *models.Portfolio) error {
	// Start a transaction
	tx := r.db.Begin()

	// Update the basic portfolio fields
	if err := tx.Save(portfolio).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update portfolio: %w", err)
	}

	// Update the Stocks association explicitly
	if err := tx.Model(portfolio).Association("Stocks").Replace(portfolio.Stocks); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update stocks for portfolio: %w", err)
	}

	// Commit the transaction
	return tx.Commit().Error
}

// UpdatePortfolioPoints update points of a portfolio specifically
func (r *PortfolioRepository) UpdatePortfolioPoints(portfolioID uint, newPoints int) error {
	err := r.db.Model(&models.Portfolio{}).
		Where("id = ?", portfolioID).
		Update("points", newPoints).Error
	if err != nil {
		return fmt.Errorf("failed to update portfolio points: %v", err)
	}
	return nil
}

// DeletePortfolio deletes a portfolio by its ID.
func (r *PortfolioRepository) DeletePortfolio(portfolioID uint) error {
	return r.db.Delete(&models.Portfolio{}, portfolioID).Error
}

// Helper Functions
func (r *PortfolioRepository) UserExists(userID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&models.User{}).Where("id = ?", userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PortfolioRepository) LeagueExists(leagueID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&models.League{}).Where("id = ?", leagueID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PortfolioRepository) GetPortfolioPointsHistoryEntry(portfolioID uint) ([]models.PortfolioPointsHistory, error) {
	var history []models.PortfolioPointsHistory
	err := r.db.
		Where("portfolio_id = ?", portfolioID).
		Order("recorded_at DESC"). // Sort by most recent first
		Find(&history).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve portfolio points history: %v", err)
	}
	return history, nil
}

func (r *PortfolioRepository) LogPortfolioPointsChange(portfolioID uint, newPoints int) error {
	historyEntry := models.PortfolioPointsHistory{
		PortfolioID: portfolioID,
		Points:      newPoints,
		RecordedAt:  time.Now(),
	}

	err := r.db.Create(&historyEntry).Error
	if err != nil {
		return fmt.Errorf("failed to log portfolio points change: %v", err)
	}
	return nil
}
