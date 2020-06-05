package state

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/byuoitav/av-control-api/api/mock"
	"github.com/google/go-cmp/cmp"
)

var getPowerTest = []stateTest{
	stateTest{
		name:          "simple",
		deviceService: mock.SimpleRoom{},
		env:           "default",
		resp: generateActionsResponse{
			Actions: []action{
				action{
					ID:  "ITB-1101-D1",
					Req: newRequest(http.MethodGet, "http://ITB-1101-CP1.byu.edu/ITB-1101-D1.av/GetPower"),
				},
			},
			ExpectedUpdates: 1,
		},
	},
}

func TestGetPower(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, tt := range getPowerTest {
		t.Run(tt.name, func(t *testing.T) {
			room, err := tt.deviceService.Room(ctx, tt.room)
			if err != nil {
				t.Errorf("unable to get room: %s", err)
			}

			var get getPower
			resp := get.GenerateActions(ctx, room, tt.env)

			if diff := cmp.Diff(tt.resp, resp); diff != "" {
				t.Errorf("generated incorrect actions (-want, +got):\n%s", diff)
			}
		})
	}
}
