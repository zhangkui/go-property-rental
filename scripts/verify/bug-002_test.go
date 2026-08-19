package verify

import (
	"context"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
	"reflect"
	"testing"
)

type facilityStoreStub struct {
	current    []entity.Facility
	replaced   []string
	propertyID string
}

func (s *facilityStoreStub) List(context.Context) ([]entity.Facility, error) { return nil, nil }
func (s *facilityStoreStub) Create(context.Context, entity.Facility) error   { return nil }
func (s *facilityStoreStub) ForProperty(context.Context, string) ([]entity.Facility, error) {
	return s.current, nil
}
func (s *facilityStoreStub) ReplacePropertyFacilities(_ context.Context, id string, ids []string) error {
	s.propertyID = id
	s.replaced = append([]string(nil), ids...)
	return nil
}
func TestBug002_BusinessRegression(t *testing.T) {
	t.Run("replacement discards stale facilities", func(t *testing.T) {
		s := &facilityStoreStub{current: []entity.Facility{{ID: "old-a"}, {ID: "old-b"}}}
		err := (service.FacilityService{Repo: s}).Replace(context.Background(), "property-1", []string{"new-c"})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(s.replaced, []string{"new-c"}) {
			t.Fatalf("replacement retained stale facilities: %#v", s.replaced)
		}
	})
	t.Run("empty replacement clears all facilities", func(t *testing.T) {
		s := &facilityStoreStub{current: []entity.Facility{{ID: "old-a"}}}
		err := (service.FacilityService{Repo: s}).Replace(context.Background(), "property-1", nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(s.replaced) != 0 {
			t.Fatalf("empty replacement retained facilities: %#v", s.replaced)
		}
	})
}
