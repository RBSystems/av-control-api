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

type CreateVideoSwitcherFunc func(context.Context, string) (VideoSwitcher, error)

// TODO should we just make an explicit input/output struct that these return in their http calls?
func CreateVideoSwitcherServer(create CreateVideoSwitcherFunc) Server {
	e := newEchoServer()
	m := &sync.Map{}

	vs := func(ctx context.Context, addr string) (VideoSwitcher, error) {
		if vs, ok := m.Load(addr); ok {
			return vs.(VideoSwitcher), nil
		}

		vs, err := create(ctx, addr)
		if err != nil {
			return nil, err
		}

		m.Store(addr, vs)
		return vs, nil
	}

	dev := func(ctx context.Context, addr string) (Device, error) {
		return vs(ctx, addr)
	}

	addDeviceRoutes(e, dev)
	addVideoSwitcherRoutes(e, vs)

	return wrapEchoServer(e)
}

func addVideoSwitcherRoutes(e *echo.Echo, create CreateVideoSwitcherFunc) {
	e.GET("/:address/output/:output/input", func(c echo.Context) error {
		addr := c.Param("address")
		out := c.Param("output")
		switch {
		case len(addr) == 0:
			return c.String(http.StatusBadRequest, "must include the address of the video switcher")
		case len(out) == 0:
			return c.String(http.StatusBadRequest, "must include an output port for the video switcher")
		}

		vs, err := create(c.Request().Context(), addr)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		in, err := vs.GetInputByOutput(c.Request().Context(), out)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, input{Input: fmt.Sprintf("%v:%v", in, out)})
	})

	e.GET("/:address/output/:output/input/:input", func(c echo.Context) error {
		addr := c.Param("address")
		out := c.Param("output")
		in := c.Param("input")
		switch {
		case len(addr) == 0:
			return c.String(http.StatusBadRequest, "must include the address of the video switcher")
		case len(out) == 0:
			return c.String(http.StatusBadRequest, "must include an output port")
		case len(in) == 0:
			return c.String(http.StatusBadRequest, "must include an input portr")
		}

		vs, err := create(c.Request().Context(), addr)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		if err := vs.SetInputByOutput(c.Request().Context(), out, in); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, input{Input: fmt.Sprintf("%v:%v", in, out)})
	})
}
