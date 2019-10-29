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
	GetActiveSignal(ctx context.Context) (bool, error)

	SetPower(ctx context.Context, power string) error
	SetBlanked(ctx context.Context, blanked bool) error
	SetInput(ctx context.Context, input string) error
}

type Display interface {
	Device
	display
}

type CreateDisplayFunc func(string) Display

func CreateDisplayServer(create CreateDisplayFunc) Server {
	e := newEchoServer()
	m := &sync.Map{}

	disp := func(addr string) Display {
		if disp, ok := m.Load(addr); ok {
			return disp.(Display)
		}

		disp := create(addr)
		m.Store(addr, disp)
		return disp
	}

	dev := func(addr string) Device {
		return disp(addr)
	}

	addDeviceRoutes(e, dev)
	addDisplayRoutes(e, disp)

	return wrapEchoServer(e)
}

func addDisplayRoutes(e *echo.Echo, create CreateDisplayFunc) {
	// power
	e.GET("/:address/power", func(c echo.Context) error {
		addr := c.Param("address")
		if len(addr) == 0 {
			return c.String(http.StatusBadRequest, "must include the address of the display")
		}

		disp := create(addr)
		power, err := disp.GetPower(c.Request().Context())
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, Power{Power: power})
	})

	e.GET("/:address/power/:power", func(c echo.Context) error {
		addr := c.Param("address")
		power := c.Param("power")
		switch {
		case len(addr) == 0:
			return c.String(http.StatusBadRequest, "must include the address of the display")
		case len(power) == 0:
			return c.String(http.StatusBadRequest, "must include a power state to set")
		}

		disp := create(addr)
		if err := disp.SetPower(c.Request().Context(), power); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, Power{Power: power})
	})

	// blanked
	e.GET("/:address/blanked", func(c echo.Context) error {
		addr := c.Param("address")
		if len(addr) == 0 {
			return c.String(http.StatusBadRequest, "must include the address of the display")
		}

		disp := create(addr)
		blanked, err := disp.GetBlanked(c.Request().Context())
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, Blanked{Blanked: blanked})
	})

	e.GET("/:address/blanked/:blanked", func(c echo.Context) error {
		addr := c.Param("address")
		blanked, err := strconv.ParseBool(c.Param("blanked"))
		switch {
		case len(addr) == 0:
			return c.String(http.StatusBadRequest, "must include the address of the display")
		case err != nil:
			return c.String(http.StatusBadRequest, err.Error())
		}

		disp := create(addr)
		if err := disp.SetBlanked(c.Request().Context(), blanked); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, Blanked{Blanked: blanked})
	})

	// input
	e.GET("/:address/input", func(c echo.Context) error {
		addr := c.Param("address")
		if len(addr) == 0 {
			return c.String(http.StatusBadRequest, "must include the address of the display")
		}

		disp := create(addr)
		input, err := disp.GetInput(c.Request().Context())
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, Input{Input: input})
	})

	e.GET("/:address/input/:input", func(c echo.Context) error {
		addr := c.Param("address")
		input := c.Param("input")
		switch {
		case len(addr) == 0:
			return c.String(http.StatusBadRequest, "must include the address of the display")
		case len(input) == 0:
			return c.String(http.StatusBadRequest, "must include a input to set")
		}

		disp := create(addr)
		if err := disp.SetInput(c.Request().Context(), input); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, Input{Input: input})
	})

	// active signal
	e.GET("/:address/activesignal", func(c echo.Context) error {
		addr := c.Param("address")
		if len(addr) == 0 {
			return c.String(http.StatusBadRequest, "must include the address of the display")
		}

		disp := create(addr)
		asignal, err := disp.GetActiveSignal(c.Request().Context())
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, ActiveSignal{ActiveSignal: asignal})
	})
}
