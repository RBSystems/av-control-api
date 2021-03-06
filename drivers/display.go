package drivers

import (
	"context"
	"net/http"
	"strconv"
	"sync"

	"github.com/labstack/echo"
)

type display interface {
	GetPower(ctx context.Context) (string, error)
	GetBlanked(ctx context.Context) (bool, error)
	GetInput(ctx context.Context) (string, error)
	GetActiveSignal(ctx context.Context, port string) (bool, error)

	SetPower(ctx context.Context, power string) error
	SetBlanked(ctx context.Context, blanked bool) error
	SetInput(ctx context.Context, input string) error
}

type Display interface {
	Device
	display
}

type CreateDisplayFunc func(context.Context, string) (Display, error)

func CreateDisplayServer(create CreateDisplayFunc) (Server, error) {
	e := newEchoServer()
	m := &sync.Map{}

	disp := func(ctx context.Context, addr string) (Display, error) {
		if disp, ok := m.Load(addr); ok {
			return disp.(Display), nil
		}

		disp, err := create(ctx, addr)
		if err != nil {
			return nil, err
		}
		m.Store(addr, disp)
		return disp, nil
	}

	dev := func(ctx context.Context, addr string) (Device, error) {
		return disp(ctx, addr)
	}

	addDeviceRoutes(e, dev)
	addDisplayRoutes(e, disp)

	return wrapEchoServer(e), nil
}

func addDisplayRoutes(e *echo.Echo, create CreateDisplayFunc) {
	// power
	e.GET("/:address/power", func(c echo.Context) error {
		addr := c.Param("address")
		if len(addr) == 0 {
			return c.String(http.StatusBadRequest, "must include the address of the display")
		}

		disp, err := create(c.Request().Context(), addr)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		pow, err := disp.GetPower(c.Request().Context())
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, power{Power: pow})
	})

	e.GET("/:address/power/:power", func(c echo.Context) error {
		addr := c.Param("address")
		pow := c.Param("power")
		switch {
		case len(addr) == 0:
			return c.String(http.StatusBadRequest, "must include the address of the display")
		case len(pow) == 0:
			return c.String(http.StatusBadRequest, "must include a power state to set")
		}

		disp, err := create(c.Request().Context(), addr)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		if err := disp.SetPower(c.Request().Context(), pow); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, power{Power: pow})
	})

	// blanked
	e.GET("/:address/blanked", func(c echo.Context) error {
		addr := c.Param("address")
		if len(addr) == 0 {
			return c.String(http.StatusBadRequest, "must include the address of the display")
		}

		disp, err := create(c.Request().Context(), addr)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		blank, err := disp.GetBlanked(c.Request().Context())
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, blanked{Blanked: blank})
	})

	e.GET("/:address/blanked/:blanked", func(c echo.Context) error {
		addr := c.Param("address")
		blank, err := strconv.ParseBool(c.Param("blanked"))
		switch {
		case len(addr) == 0:
			return c.String(http.StatusBadRequest, "must include the address of the display")
		case err != nil:
			return c.String(http.StatusBadRequest, err.Error())
		}

		disp, err := create(c.Request().Context(), addr)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		if err := disp.SetBlanked(c.Request().Context(), blank); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, blanked{Blanked: blank})
	})

	// input
	e.GET("/:address/input", func(c echo.Context) error {
		addr := c.Param("address")
		if len(addr) == 0 {
			return c.String(http.StatusBadRequest, "must include the address of the display")
		}

		disp, err := create(c.Request().Context(), addr)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		in, err := disp.GetInput(c.Request().Context())
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, input{Input: in})
	})

	e.GET("/:address/input/:input", func(c echo.Context) error {
		addr := c.Param("address")
		in := c.Param("input")
		switch {
		case len(addr) == 0:
			return c.String(http.StatusBadRequest, "must include the address of the display")
		case len(in) == 0:
			return c.String(http.StatusBadRequest, "must include a input to set")
		}

		disp, err := create(c.Request().Context(), addr)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		if err := disp.SetInput(c.Request().Context(), in); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, input{Input: in})
	})

	// active signal
	e.GET("/:address/activesignal/:port", func(c echo.Context) error {
		addr := c.Param("address")
		port := c.Param("port")
		if len(addr) == 0 {
			return c.String(http.StatusBadRequest, "must include the address of the display")
		}

		disp, err := create(c.Request().Context(), addr)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		asignal, err := disp.GetActiveSignal(c.Request().Context(), port)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, activeSignal{ActiveSignal: asignal})
	})
}
