package state

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/byuoitav/av-control-api/api/mock"
	"github.com/google/go-cmp/cmp"
)

var getVolumeTest = []stateTest{
	stateTest{
		name:          "simpleSeparateInput",
		deviceService: mock.SimpleSeparateInput{},
		env:           "default",
		resp: generateActionsResponse{
			Actions: []action{
				action{
					ID: "ITB-1101-AMP1",
					Req: &http.Request{
						Method: http.MethodGet,
						URL:    urlParse("http://ITB-1101-AMP1.av/GetVolume"),
					},
				},
			},
			ExpectedUpdates: 1,
		},
	},
}

func TestGetVolume(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, tt := range getVolumeTest {
		t.Run(tt.name, func(t *testing.T) {
			room, err := tt.deviceService.Room(ctx, tt.room)
			if err != nil {
				t.Errorf("unable to get room: %s", err)
			}

			var get getVolume
			resp := get.GenerateActions(ctx, room, tt.env)

			if !cmp.Equal(resp, tt.resp) {
				t.Errorf("generated incorrect actions:\n\tgot %+v\n\texpected: %+v", resp, tt.resp)
			}
		})
	}
}
