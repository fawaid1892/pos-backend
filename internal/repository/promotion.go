package repository

import (
	"errors"
	"time"

	"pos-multi-branch/backend/internal/database"
	"pos-multi-branch/backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─── Promotion ───

type ListPromotionsParams struct {
	Active   *bool
	Expired  *bool
	BranchID *uuid.UUID
	Type     string
	Limit    int
	Offset   int
}

func ListPromotions(p ListPromotionsParams) ([]model.Promotion, error) {
	query := database.DB.Model(&model.Promotion{}).Where("deleted_at IS NULL")

	if p.Active != nil && *p.Active {
		now := time.Now()
		query = query.Where("start_date <= ? AND end_date >= ? AND is_active = true", now, now)
	}
	if p.Expired != nil && *p.Expired {
		now := time.Now()
		query = query.Where("end_date < ?", now)
	}
	if p.BranchID != nil {
		// A promotion applies to a branch if:
		// scope = all, OR
		// scope = province AND province_id matches branch's province, OR
		// scope = city AND city_id matches branch's city, OR
		// scope = selected AND branch is in promotion_branches
		branchID := *p.BranchID

		// Subquery: promotions whose scope = 'all'
		// OR scope = 'province' with matching province_id (need branch's province)
		// OR scope = 'city' with matching city_id (need branch's city)
		// OR scope = 'selected' with join to promotion_branches
		//
		// To avoid complex subqueries, use a raw JOIN approach:
		// SELECT p.* FROM promotions p
		// WHERE p.deleted_at IS NULL
		//   AND (
		//     p.scope = 'all'
		//     OR (p.scope = 'selected' AND EXISTS (
		//       SELECT 1 FROM promotion_branches pb WHERE pb.promotion_id = p.id AND pb.branch_id = ?
		//     ))
		//     OR (p.scope = 'province' AND p.province_id = (SELECT province FROM branches WHERE id = ?))
		//     OR (p.scope = 'city' AND p.city_id = (SELECT city FROM branches WHERE id = ?))
		//   )
		query = query.Where(
			`(scope = 'all') OR
			 (scope = 'selected' AND EXISTS (
			   SELECT 1 FROM promotion_branches pb WHERE pb.promotion_id = promotions.id AND pb.branch_id = ?
			 )) OR
			 (scope = 'province' AND province_id = (SELECT province_code FROM branches WHERE id = ?)) OR
			 (scope = 'city' AND city_id = (SELECT city_code FROM branches WHERE id = ?))`,
			branchID, branchID, branchID,
		)
	}
	if p.Type != "" {
		query = query.Where("type = ?", p.Type)
	}
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 20
	}

	var promos []model.Promotion
	err := query.Order("created_at DESC").Limit(p.Limit).Offset(p.Offset).Find(&promos).Error
	return promos, err
}

func GetPromotionByID(id uuid.UUID) (*model.Promotion, error) {
	p := &model.Promotion{}
	err := database.DB.Preload("Branches").Where("id = ? AND deleted_at IS NULL", id).First(p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}

func CreatePromotion(req model.CreatePromotionRequest) (*model.Promotion, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	scope := req.Scope
	if scope == "" {
		scope = "selected"
	}

	p := &model.Promotion{
		Name:          req.Name,
		Type:          req.Type,
		Code:          req.Code,
		DiscountValue: req.DiscountValue,
		DiscountType:  req.DiscountType,
		SkuTarget:     req.SkuTarget,
		QtyMin:        req.QtyMin,
		QtyFree:       req.QtyFree,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		Scope:         scope,
		ProvinceID:    req.ProvinceID,
		CityID:        req.CityID,
		IsActive:      isActive,
		MaxUses:       req.MaxUses,
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(p).Error; err != nil {
			return err
		}

		// If scope=selected, create PromotionBranch entries
		if scope == "selected" && len(req.BranchIDs) > 0 {
			for _, bidStr := range req.BranchIDs {
				bid, err := uuid.Parse(bidStr)
				if err != nil {
					return errors.New("invalid branch_id in branch_ids: " + err.Error())
				}
				pb := model.PromotionBranch{
					PromotionID: p.ID,
					BranchID:    bid,
				}
				if err := tx.Create(&pb).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Reload with branches
	database.DB.Preload("Branches").First(p, p.ID)
	return p, nil
}

func UpdatePromotion(id uuid.UUID, req model.UpdatePromotionRequest) (*model.Promotion, error) {
	updates := map[string]interface{}{}

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.Code != nil {
		updates["code"] = *req.Code
	}
	if req.DiscountValue != nil {
		updates["discount_value"] = *req.DiscountValue
	}
	if req.DiscountType != nil {
		updates["discount_type"] = *req.DiscountType
	}
	if req.SkuTarget != nil {
		updates["sku_target"] = *req.SkuTarget
	}
	if req.QtyMin != nil {
		updates["qty_min"] = *req.QtyMin
	}
	if req.QtyFree != nil {
		updates["qty_free"] = *req.QtyFree
	}
	if req.StartDate != nil {
		updates["start_date"] = *req.StartDate
	}
	if req.EndDate != nil {
		updates["end_date"] = *req.EndDate
	}
	if req.Scope != nil {
		updates["scope"] = *req.Scope
	}
	if req.ProvinceID != nil {
		updates["province_id"] = *req.ProvinceID
	}
	if req.CityID != nil {
		updates["city_id"] = *req.CityID
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.MaxUses != nil {
		updates["max_uses"] = *req.MaxUses
	}

	if len(updates) == 0 && len(req.BranchIDs) == 0 {
		return nil, errors.New("no fields to update")
	}

	p := &model.Promotion{}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Update the promotion fields
		if len(updates) > 0 {
			if err := tx.Model(p).Where("id = ? AND deleted_at IS NULL", id).Updates(updates).Error; err != nil {
				return err
			}
		}

		// If BranchIDs provided, update the join table
		if req.BranchIDs != nil {
			// Delete old branches
			if err := tx.Where("promotion_id = ?", id).Delete(&model.PromotionBranch{}).Error; err != nil {
				return err
			}
			// Insert new branches
			for _, bidStr := range req.BranchIDs {
				bid, err := uuid.Parse(bidStr)
				if err != nil {
					return errors.New("invalid branch_id in branch_ids: " + err.Error())
				}
				pb := model.PromotionBranch{
					PromotionID: id,
					BranchID:    bid,
				}
				if err := tx.Create(&pb).Error; err != nil {
					return err
				}
			}
		}

		// Reload the promotion
		if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(p).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil // will be caught by nil check below
			}
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if p.ID == uuid.Nil {
		return nil, nil
	}

	// Preload branches
	database.DB.Preload("Branches").First(p, p.ID)
	return p, nil
}

func SoftDeletePromotion(id uuid.UUID) error {
	result := database.DB.Where("id = ? AND deleted_at IS NULL", id).Delete(&model.Promotion{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("promotion not found")
	}
	return nil
}

func GetActivePromotions() ([]model.Promotion, error) {
	now := time.Now()
	var promos []model.Promotion
	err := database.DB.Where("deleted_at IS NULL AND is_active = true AND start_date <= ? AND end_date >= ?", now, now).
		Order("created_at DESC").
		Find(&promos).Error
	return promos, err
}

// ListPromotionsByBranch returns active promotions that apply to a specific branch.
func ListPromotionsByBranch(branchID uuid.UUID) ([]model.Promotion, error) {
	now := time.Now()
	var promos []model.Promotion

	err := database.DB.Model(&model.Promotion{}).
		Where("deleted_at IS NULL AND is_active = true AND start_date <= ? AND end_date >= ?", now, now).
		Where(
			`(scope = 'all') OR
			 (scope = 'selected' AND EXISTS (
			   SELECT 1 FROM promotion_branches pb WHERE pb.promotion_id = promotions.id AND pb.branch_id = ?
			 )) OR
			 (scope = 'province' AND province_id = (SELECT province_code FROM branches WHERE id = ?)) OR
			 (scope = 'city' AND city_id = (SELECT city_code FROM branches WHERE id = ?))`,
			branchID, branchID, branchID,
		).
		Order("created_at DESC").
		Find(&promos).Error

	if promos == nil {
		promos = []model.Promotion{}
	}
	return promos, err
}

func ValidateVoucher(code string, branchID *uuid.UUID, total float64) (*model.ValidateVoucherResponse, error) {
	now := time.Now()

	var promo model.Promotion
	err := database.DB.Where("code = ? AND deleted_at IS NULL AND is_active = true AND start_date <= ? AND end_date >= ?",
		code, now, now).First(&promo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.ValidateVoucherResponse{Valid: false, Error: "Voucher tidak ditemukan atau sudah tidak aktif"}, nil
		}
		return nil, err
	}

	// Check max uses
	if promo.MaxUses > 0 && promo.CurrentUses >= promo.MaxUses {
		return &model.ValidateVoucherResponse{Valid: false, Error: "Voucher sudah mencapai batas maksimal pemakaian"}, nil
	}

	// Check branch — using multi-branch scope logic
	if branchID != nil {
		applies, err := promotionAppliesToBranch(promo.ID, *branchID)
		if err != nil {
			return nil, err
		}
		if !applies {
			return &model.ValidateVoucherResponse{Valid: false, Error: "Voucher tidak berlaku untuk cabang ini"}, nil
		}
	}

	// Check min_purchase
	if promo.Type == "min_purchase" && total < float64(promo.QtyMin) {
		return &model.ValidateVoucherResponse{Valid: false, Error: "Total pembelian belum mencapai minimum"}, nil
	}

	return &model.ValidateVoucherResponse{
		Valid:         true,
		DiscountValue: promo.DiscountValue,
		DiscountType:  promo.DiscountType,
		PromotionName: promo.Name,
	}, nil
}

// promotionAppliesToBranch checks if a promotion applies to a given branch based on scope.
func promotionAppliesToBranch(promotionID uuid.UUID, branchID uuid.UUID) (bool, error) {
	var promo model.Promotion
	err := database.DB.Where("id = ?", promotionID).First(&promo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	switch promo.Scope {
	case "all":
		return true, nil
	case "selected":
		var count int64
		database.DB.Model(&model.PromotionBranch{}).
			Where("promotion_id = ? AND branch_id = ?", promotionID, branchID).
			Count(&count)
		return count > 0, nil
	case "province":
		var branch model.Branch
		if err := database.DB.Where("id = ?", branchID).First(&branch).Error; err != nil {
			return false, err
		}
		return branch.ProvinceCode == promo.ProvinceID, nil
	case "city":
		var branch model.Branch
		if err := database.DB.Where("id = ?", branchID).First(&branch).Error; err != nil {
			return false, err
		}
		return branch.CityCode == promo.CityID, nil
	default:
		return false, nil
	}
}

func IncrementPromotionUses(id uuid.UUID) error {
	return database.DB.Model(&model.Promotion{}).Where("id = ?", id).
		UpdateColumn("current_uses", gorm.Expr("current_uses + 1")).Error
}
