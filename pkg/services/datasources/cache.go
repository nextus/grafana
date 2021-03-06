package datasources

import (
	"fmt"
	"time"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/infra/localcache"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/registry"
)

type CacheService interface {
	GetDatasource(datasourceID int64, user *models.SignedInUser, skipCache bool) (*models.DataSource, error)
}

type CacheServiceImpl struct {
	Bus          bus.Bus                  `inject:""`
	CacheService *localcache.CacheService `inject:""`
}

func init() {
	registry.Register(&registry.Descriptor{
		Name:         "DatasourceCacheService",
		Instance:     &CacheServiceImpl{},
		InitPriority: registry.Low,
	})
}

func (dc *CacheServiceImpl) Init() error {
	return nil
}

func (dc *CacheServiceImpl) GetDatasource(datasourceID int64, user *models.SignedInUser, skipCache bool) (*models.DataSource, error) {
	cacheKey := fmt.Sprintf("ds-%d", datasourceID)

	if !skipCache {
		if cached, found := dc.CacheService.Get(cacheKey); found {
			ds := cached.(*models.DataSource)
			if ds.OrgId == user.OrgId {
				return ds, nil
			}
		}
	}

	query := models.GetDataSourceByIdQuery{Id: datasourceID, OrgId: user.OrgId}
	if err := dc.Bus.Dispatch(&query); err != nil {
		return nil, err
	}

	dc.CacheService.Set(cacheKey, query.Result, time.Second*5)
	return query.Result, nil
}
