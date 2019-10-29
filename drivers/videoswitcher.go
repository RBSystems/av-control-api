package drivers

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/labstack/echo"
)

type videoSwitcher interface {
	// TODO notes about being 1 indexed

	GetInputByOutput(ctx context.Context, output string) (string, error)
	SetInputByOutput(ctx context.Context, output, input string) error

	// TODO active input ?
}

type VideoSwitcher interface {
	Device
	videoSwitcher
}

type CreateVideoSwitcherFunc func(string) VideoSwitcher

// TODO should we just make an explicit input/output struct that these return in their http calls?
func CreateVideoSwitcherServer(create CreateVideoSwitcherFunc) Server {
	e := newEchoServer()
	m := &sync.Map{}

	vs := func(addr string) VideoSwitcher {
		if vs, ok := m.Load(addr); ok {
			return vs.(VideoSwitcher)
		}

		vs := create(addr)
		m.Store(addr, vs)
		return vs
	}

	dev := func(addr string) Device {
		return vs(addr)
	}

	addDeviceRoutes(e, dev)
	addVideoSwitcherRoutes(e, vs)

	return wrapEchoServer(e)
}

func addVideoSwitcherRoutes(e *echo.Echo, create CreateVideoSwitcherFunc) {
	e.GET("/:address/output/:output/input", func(c echo.Context) error {
		addr := c.Param("address")
		output := c.Param("output")
		switch {
		case len(addr) == 0:
			return c.String(http.StatusBadRequest, "must include the address of the video switcher")
		case len(output) == 0:
			return c.String(http.StatusBadRequest, "must include an output port for the video switcher")
		}

		vs := create(addr)
		input, err := vs.GetInputByOutput(c.Request().Context(), output)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, Input{Input: fmt.Sprintf("%v:%v", input, output)})
	})

	e.GET("/:address/output/:output/input/:input", func(c echo.Context) error {
		addr := c.Param("address")
		output := c.Param("output")
		input := c.Param("input")
		switch {
		case len(addr) == 0:
			return c.String(http.StatusBadRequest, "must include the address of the video switcher")
		case len(output) == 0:
			return c.String(http.StatusBadRequest, "must include an output port")
		case len(input) == 0:
			return c.String(http.StatusBadRequest, "must include an input portr")
		}

		vs := create(addr)
		if err := vs.SetInputByOutput(c.Request().Context(), output, input); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, Input{Input: fmt.Sprintf("%v:%v", input, output)})
	})
}
