package service

import (
	"encoding/json"
	"strings"

	"github.com/deposist/s-ui-x/database"
	"github.com/deposist/s-ui-x/database/model"
	"github.com/deposist/s-ui-x/util"
	"github.com/deposist/s-ui-x/util/common"

	"gorm.io/gorm"
)

type TlsService struct {
	InboundService
	ServicesService
	Runtime *Runtime
}

func (s *TlsService) GetAll() ([]model.Tls, error) {
	db := database.GetDB()
	tlsConfig := []model.Tls{}
	err := db.Model(model.Tls{}).Where("id > 0").Scan(&tlsConfig).Error
	if err != nil {
		return nil, err
	}

	return tlsConfig, nil
}

// Save persists a TLS change inside tx and returns the ids of inbounds and
// services that reference the edited TLS row. The caller hot-reloads them
// after the transaction commits, so the running core picks up the new
// certificate from committed state without a full restart (and without the
// pre-commit reload window where a rollback would leave the core ahead of the
// database).
func (s *TlsService) Save(tx *gorm.DB, action string, data json.RawMessage, hostname string) (inboundIds []uint, serviceIds []uint, err error) {
	switch action {
	case "new", "edit":
		var tls model.Tls
		err = json.Unmarshal(data, &tls)
		if err != nil {
			return nil, nil, err
		}
		if err = normalizeRealityServerTLS(&tls); err != nil {
			return nil, nil, err
		}
		err = tx.Save(&tls).Error
		if err != nil {
			return nil, nil, err
		}
		if action != "edit" {
			return nil, nil, nil
		}
		var inbounds []model.Inbound
		err = tx.Model(model.Inbound{}).Preload("Tls").Where("tls_id = ?", tls.Id).Find(&inbounds).Error
		if err != nil {
			return nil, nil, err
		}
		if len(inbounds) > 0 {
			err = s.ClientService.UpdateLinksByInboundChange(tx, &inbounds, hostname, "")
			if err != nil {
				return nil, nil, err
			}
			for _, inbound := range inbounds {
				inboundIds = append(inboundIds, inbound.Id)
			}
			err = s.InboundService.UpdateOutJsons(tx, inboundIds, hostname)
			if err != nil {
				return nil, nil, common.NewError("unable to update out_json of inbounds: ", err.Error())
			}
		}
		err = tx.Model(model.Service{}).Select("id").Where("tls_id = ?", tls.Id).Scan(&serviceIds).Error
		if err != nil {
			return nil, nil, err
		}
		return inboundIds, serviceIds, nil
	case "del":
		var id uint
		err = json.Unmarshal(data, &id)
		if err != nil {
			return nil, nil, err
		}
		var inboundCount int64
		err = tx.Model(model.Inbound{}).Where("tls_id = ?", id).Count(&inboundCount).Error
		if err != nil {
			return nil, nil, err
		}
		var serviceCount int64
		err = tx.Model(model.Service{}).Where("tls_id = ?", id).Count(&serviceCount).Error
		if err != nil {
			return nil, nil, err
		}
		if inboundCount > 0 || serviceCount > 0 {
			return nil, nil, common.NewError("tls in use")
		}
		err = tx.Where("id = ?", id).Delete(model.Tls{}).Error
		if err != nil {
			return nil, nil, err
		}
	}

	return nil, nil, nil
}

// normalizeRealityServerTLS repairs malformed legacy Reality values before they
// are persisted and applied to sing-box. Invalid Short IDs are discarded; an
// empty list is valid and avoids aborting the entire core restart.
func normalizeRealityServerTLS(tls *model.Tls) error {
	var server map[string]any
	if err := json.Unmarshal(tls.Server, &server); err != nil {
		return err
	}
	reality, ok := server["reality"].(map[string]any)
	if !ok {
		return nil
	}
	if ids, ok := reality["short_id"].([]any); ok {
		valid := make([]string, 0, len(ids))
		for _, value := range ids {
			if id, ok := value.(string); ok && util.ValidRealityShortID(id) {
				valid = append(valid, id)
			}
		}
		if len(valid) != len(ids) {
			reality["short_id"] = valid
		}
	}
	if name, _ := server["server_name"].(string); strings.TrimSpace(name) == "" {
		if handshake, ok := reality["handshake"].(map[string]any); ok {
			if host, ok := handshake["server"].(string); ok && host != "" {
				server["server_name"] = host
			}
		}
	}
	updated, err := json.Marshal(server)
	if err != nil {
		return err
	}
	tls.Server = updated
	return nil
}
