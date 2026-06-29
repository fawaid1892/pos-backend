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
		query = query.Where("branch_id = ? OR branch_id IS NULL", *p.BranchID)
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
	err := database.DB.Where("id = ? AND deleted_at IS NULL", id).First(p).Error
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
		BranchID:      req.BranchID,
		IsActive:      isActive,
		MaxUses:       req.MaxUses,
	}
	err := database.DB.Create(p).Error
	if err != nil {
		return nil, err
	}
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
	if req.BranchID != nil {
		updates["branch_id"] = *req.BranchID
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.MaxUses != nil {
		updates["max_uses"] = *req.MaxUses
	}

	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	p := &model.Promotion{}
	err := database.DB.Model(p).Where("id = ? AND deleted_at IS NULL", id).Updates(updates).First(p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
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

	// Check branch
	if promo.BranchID != nil && branchID != nil {
		if *promo.BranchID != *branchID {
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

func IncrementPromotionUses(id uuid.UUID) error {
	return database.DB.Model(&model.Promotion{}).Where("id = ?", id).
		UpdateColumn("current_uses", gorm.Expr("current_uses + 1")).Error
}
