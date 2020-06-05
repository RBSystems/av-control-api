package state

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/byuoitav/av-control-api/api"
	"github.com/byuoitav/av-control-api/api/graph"
	gonum "gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/traverse"
)

type getInput struct{}

// GenerateActions makes an assumption that GetInput and GetInputByBlock will not ever be on the same device
func (i *getInput) GenerateActions(ctx context.Context, room []api.Device, env string) generateActionsResponse {
	var resp generateActionsResponse
	responses := make(chan actionResponse)
	g := graph.NewGraph(room, "video")

	outputs := graph.Leaves(g)

	t := graph.Transpose(g)
	inputs := graph.Leaves(t)

	paths := path.DijkstraAllPaths(t)

	for _, output := range outputs {
		var actsForOutput []action
		var errsForOutput []api.DeviceStateError

		for _, input := range inputs {
			path := graph.PathFromTo(t, &paths, output.Device.ID, input.Device.ID)
			if len(path) == 0 {
				continue
			}

			acts, errs := i.generateActionsForPath(ctx, path, env, responses)
			fmt.Printf("len acts: %d\n", len(acts))
			actsForOutput = append(actsForOutput, acts...)
			errsForOutput = append(errsForOutput, errs...)
		}

		if len(errsForOutput) == 0 && len(actsForOutput) != 0 {
			resp.ExpectedUpdates++
			resp.Actions = append(resp.Actions, actsForOutput...)
		}

		resp.Errors = append(resp.Errors, errsForOutput...)
	}

	fmt.Printf("unexpected: %d\n", resp.ExpectedUpdates)

	if resp.ExpectedUpdates == 0 {
		fmt.Printf("no expected updates\n")
		return generateActionsResponse{}
	}

	// combine identical actions
	resp.Actions = uniqueActions(resp.Actions)

	if len(resp.Actions) > 0 {
		go i.handleResponses(responses, len(resp.Actions), resp.ExpectedUpdates, t, &paths, outputs, inputs)
	}

	return resp
}

func (i *getInput) generateActionsForPath(ctx context.Context, path graph.Path, env string, resps chan actionResponse) ([]action, []api.DeviceStateError) {
	var acts []action
	var errs []api.DeviceStateError
	for i := range path {
		switch i {
		case 0:
			// the edge leaving the output
			url, order, err := getCommand(*path[i].Src.Device, "GetAVInput", env)
			switch {
			case errors.Is(err, errCommandNotFound), errors.Is(err, errCommandEnvNotFound):
			case err != nil:
				errs = append(errs, api.DeviceStateError{
					ID:    path[i].Src.Device.ID,
					Field: "input",
					Error: err.Error(),
				})

				return acts, errs

			default:
				params := map[string]string{
					"address": path[i].Src.Address,
				}

				url, err = fillURL(url, params)
				if err != nil {
					errs = append(errs, api.DeviceStateError{
						ID:    path[i].Src.Device.ID,
						Field: "input",
						Error: fmt.Sprintf("%s (url after fill: %s)", err, url),
					})

					return acts, errs
				}

				req, err := http.NewRequest(http.MethodGet, url, nil)
				if err != nil {
					errs = append(errs, api.DeviceStateError{
						ID:    path[i].Src.Device.ID,
						Field: "input",
						Error: fmt.Sprintf("unable to build http request: %s", err),
					})

					return acts, errs
				}

				act := action{
					ID:       path[i].Src.Device.ID,
					Req:      req,
					Order:    order,
					Response: resps,
				}

				acts = append(acts, act)
			}

			url, order, err = getCommand(*path[i].Src.Device, "GetVideoInput", env)
			switch {
			case errors.Is(err, errCommandNotFound), errors.Is(err, errCommandEnvNotFound):
			case err != nil:
				errs = append(errs, api.DeviceStateError{
					ID:    path[i].Src.Device.ID,
					Field: "input",
					Error: err.Error(),
				})

				return acts, errs

			default:

				params := map[string]string{
					"address": path[i].Src.Address,
					// "port": ,
				}

				url, err = fillURL(url, params)
				if err != nil {
					errs = append(errs, api.DeviceStateError{
						ID:    path[i].Src.Device.ID,
						Field: "input",
						Error: fmt.Sprintf("%s (url after fill: %s)", err, url),
					})

					return acts, errs
				}

				req, err := http.NewRequest(http.MethodGet, url, nil)
				if err != nil {
					errs = append(errs, api.DeviceStateError{
						ID:    path[i].Src.Device.ID,
						Field: "input",
						Error: fmt.Sprintf("unable to build http request: %s", err),
					})

					return acts, errs
				}

				act := action{
					ID:       path[i].Src.Device.ID,
					Req:      req,
					Order:    order,
					Response: resps,
				}

				acts = append(acts, act)
			}

			url, order, err = getCommand(*path[i].Src.Device, "GetAudioInput", env)
			switch {
			case errors.Is(err, errCommandNotFound), errors.Is(err, errCommandEnvNotFound):
			case err != nil:
				errs = append(errs, api.DeviceStateError{
					ID:    path[i].Src.Device.ID,
					Field: "input",
					Error: err.Error(),
				})

				return acts, errs

			default:

				params := map[string]string{
					"address": path[i].Src.Address,
					// "port": ,
				}

				url, err = fillURL(url, params)
				if err != nil {
					errs = append(errs, api.DeviceStateError{
						ID:    path[i].Src.Device.ID,
						Field: "input",
						Error: fmt.Sprintf("%s (url after fill: %s)", err, url),
					})

					return acts, errs
				}

				req, err := http.NewRequest(http.MethodGet, url, nil)
				if err != nil {
					errs = append(errs, api.DeviceStateError{
						ID:    path[i].Src.Device.ID,
						Field: "input",
						Error: fmt.Sprintf("unable to build http request: %s", err),
					})

					return acts, errs
				}

				act := action{
					ID:       path[i].Src.Device.ID,
					Req:      req,
					Order:    order,
					Response: resps,
				}

				acts = append(acts, act)
			}
		default:
			// the edges between devices that aren't the output
			url, order, err := getCommand(*path[i].Src.Device, "GetAVInput", env)
			switch {
			case errors.Is(err, errCommandNotFound), errors.Is(err, errCommandEnvNotFound):
			case err != nil:
				errs = append(errs, api.DeviceStateError{
					ID:    path[i].Src.Device.ID,
					Field: "input",
					Error: err.Error(),
				})

				return acts, errs
			default:

				params := map[string]string{
					"address": path[i].Src.Address,
					// "output":  path[i-1].DstPort.Name,
				}

				url, err = fillURL(url, params)
				if err != nil {
					errs = append(errs, api.DeviceStateError{
						ID:    path[i].Src.Device.ID,
						Field: "input",
						Error: fmt.Sprintf("%s (url after fill: %s)", err, url),
					})

					return acts, errs
				}

				req, err := http.NewRequest(http.MethodGet, url, nil)
				if err != nil {
					errs = append(errs, api.DeviceStateError{
						ID:    path[i].Src.Device.ID,
						Field: "input",
						Error: fmt.Sprintf("unable to build http request: %s", err),
					})

					return acts, errs
				}

				act := action{
					ID:       path[i].Src.Device.ID,
					Req:      req,
					Order:    order,
					Response: resps,
				}

				acts = append(acts, act)
				continue
			}
			url, order, err = getCommand(*path[i].Src.Device, "GetVideoInput", env)
			switch {
			case errors.Is(err, errCommandNotFound), errors.Is(err, errCommandEnvNotFound):
			case err != nil:
				errs = append(errs, api.DeviceStateError{
					ID:    path[i].Src.Device.ID,
					Field: "input",
					Error: err.Error(),
				})

				return acts, errs
			default:

				params := map[string]string{
					"address": path[i].Src.Address,
				}

				url, err = fillURL(url, params)
				if err != nil {
					errs = append(errs, api.DeviceStateError{
						ID:    path[i].Src.Device.ID,
						Field: "input",
						Error: fmt.Sprintf("%s (url after fill: %s)", err, url),
					})

					return acts, errs
				}

				req, err := http.NewRequest(http.MethodGet, url, nil)
				if err != nil {
					errs = append(errs, api.DeviceStateError{
						ID:    path[i].Src.Device.ID,
						Field: "input",
						Error: fmt.Sprintf("unable to build http request: %s", err),
					})

					return acts, errs
				}

				act := action{
					ID:       path[i].Src.Device.ID,
					Req:      req,
					Order:    order,
					Response: resps,
				}

				acts = append(acts, act)
			}

			url, order, err = getCommand(*path[i].Src.Device, "GetAudioInput", env)
			switch {
			case errors.Is(err, errCommandNotFound), errors.Is(err, errCommandEnvNotFound):
			case err != nil:
				errs = append(errs, api.DeviceStateError{
					ID:    path[i].Src.Device.ID,
					Field: "input",
					Error: err.Error(),
				})

				return acts, errs
			default:

				params := map[string]string{
					"address": path[i].Src.Address,
				}

				url, err = fillURL(url, params)
				if err != nil {
					errs = append(errs, api.DeviceStateError{
						ID:    path[i].Src.Device.ID,
						Field: "input",
						Error: fmt.Sprintf("%s (url after fill: %s)", err, url),
					})

					return acts, errs
				}

				req, err := http.NewRequest(http.MethodGet, url, nil)
				if err != nil {
					errs = append(errs, api.DeviceStateError{
						ID:    path[i].Src.Device.ID,
						Field: "input",
						Error: fmt.Sprintf("unable to build http request: %s", err),
					})

					return acts, errs
				}

				act := action{
					ID:       path[i].Src.Device.ID,
					Req:      req,
					Order:    order,
					Response: resps,
				}

				acts = append(acts, act)
			}

			url, order, err = getCommand(*path[i].Src.Device, "GetAVInputForOutput", env)
			switch {
			case errors.Is(err, errCommandNotFound), errors.Is(err, errCommandEnvNotFound):
			case err != nil:
				errs = append(errs, api.DeviceStateError{
					ID:    path[i].Src.Device.ID,
					Field: "input",
					Error: err.Error(),
				})

				return acts, errs

			default:
				params := map[string]string{
					"address": path[i].Src.Address,
					// "port": ,
				}

				url, err = fillURL(url, params)
				if err != nil {
					errs = append(errs, api.DeviceStateError{
						ID:    path[i].Src.Device.ID,
						Field: "input",
						Error: fmt.Sprintf("%s (url after fill: %s)", err, url),
					})

					return acts, errs
				}

				req, err := http.NewRequest(http.MethodGet, url, nil)
				if err != nil {
					errs = append(errs, api.DeviceStateError{
						ID:    path[i].Src.Device.ID,
						Field: "input",
						Error: fmt.Sprintf("unable to build http request: %s", err),
					})

					return acts, errs
				}

				act := action{
					ID:       path[i].Src.Device.ID,
					Req:      req,
					Order:    order,
					Response: resps,
				}

				acts = append(acts, act)
			}

			url, order, err = getCommand(*path[i].Src.Device, "GetVideoInputForOutput", env)
			switch {
			case errors.Is(err, errCommandNotFound), errors.Is(err, errCommandEnvNotFound):
			case err != nil:
				errs = append(errs, api.DeviceStateError{
					ID:    path[i].Src.Device.ID,
					Field: "input",
					Error: err.Error(),
				})

				return acts, errs
			default:
				params := map[string]string{
					"address": path[i].Src.Address,
					// "port": ,
				}

				url, err = fillURL(url, params)
				if err != nil {
					errs = append(errs, api.DeviceStateError{
						ID:    path[i].Src.Device.ID,
						Field: "input",
						Error: fmt.Sprintf("%s (url after fill: %s)", err, url),
					})

					return acts, errs
				}

				req, err := http.NewRequest(http.MethodGet, url, nil)
				if err != nil {
					errs = append(errs, api.DeviceStateError{
						ID:    path[i].Src.Device.ID,
						Field: "input",
						Error: fmt.Sprintf("unable to build http request: %s", err),
					})

					return acts, errs
				}

				act := action{
					ID:       path[i].Src.Device.ID,
					Req:      req,
					Order:    order,
					Response: resps,
				}

				acts = append(acts, act)
			}

			url, order, err = getCommand(*path[i].Src.Device, "GetAudioInputForOutput", env)
			switch {
			case errors.Is(err, errCommandNotFound), errors.Is(err, errCommandEnvNotFound):
			case err != nil:
				errs = append(errs, api.DeviceStateError{
					ID:    path[i].Src.Device.ID,
					Field: "input",
					Error: err.Error(),
				})

				return acts, errs
			default:

				params := map[string]string{
					"address": path[i].Src.Address,
					// "port": ,
				}

				url, err = fillURL(url, params)
				if err != nil {
					errs = append(errs, api.DeviceStateError{
						ID:    path[i].Src.Device.ID,
						Field: "input",
						Error: fmt.Sprintf("%s (url after fill: %s)", err, url),
					})

					return acts, errs
				}

				req, err := http.NewRequest(http.MethodGet, url, nil)
				if err != nil {
					errs = append(errs, api.DeviceStateError{
						ID:    path[i].Src.Device.ID,
						Field: "input",
						Error: fmt.Sprintf("unable to build http request: %s", err),
					})

					return acts, errs
				}

				act := action{
					ID:       path[i].Src.Device.ID,
					Req:      req,
					Order:    order,
					Response: resps,
				}

				acts = append(acts, act)
			}
		}
	}

	return acts, errs
}

type input struct {
	Audio            *string  `json:"audio"`
	Video            *string  `json:"video"`
	CanSetSeparately *bool    `json:"canSetSeparately"`
	AvailableInputs  []string `json:"availableInputs"`
}

func (i *getInput) handleResponses(respChan chan actionResponse, expectedResps, expectedUpdates int, t *simple.DirectedGraph, paths *path.AllShortest, outputs, inputs []graph.Node) {
	if expectedResps == 0 {
		return
	}

	fmt.Printf("Expected resps: %d\n", expectedResps)
	fmt.Printf("Expected updates: %d\n", expectedUpdates)

	var resps []actionResponse
	var received int

	for resp := range respChan {
		received++
		resps = append(resps, resp)

		if received == expectedResps {
			break
		}
	}
	close(respChan)
	status := make(map[api.DeviceID][]input)
	var emptyChecker []api.DeviceID
	for _, resp := range resps {
		handleErr := func(err error) {
			resp.Errors <- api.DeviceStateError{
				ID:    resp.Action.ID,
				Field: "input",
				Error: err.Error(),
			}
		}

		if resp.Error != nil {
			handleErr(fmt.Errorf("unable to make http request: %w", resp.Error))
			continue
		}

		if resp.StatusCode/100 != 2 {
			handleErr(fmt.Errorf("%v response from driver: %s", resp.StatusCode, resp.Body))
			continue
		}

		var state input
		if err := json.Unmarshal(resp.Body, &state); err != nil {
			handleErr(fmt.Errorf("unable to parse response from driver: %w. response:\n%s", err, resp.Body))
			continue
		}

		fmt.Printf("%s input: %s\n", resp.Action.ID, resp.Body)

		// If all devices are off and we can't get any inputs for them we just need to return out
		if string(resp.Body) == "{}" {
			emptyChecker = append(emptyChecker, resp.Action.ID)
			resp.Errors <- api.DeviceStateError{
				ID:    resp.Action.ID,
				Field: "input",
				Error: fmt.Sprintf("unable to get input for %s (probably powered off)", resp.Action.ID),
			}
			resp.Updates <- OutputStateUpdate{}
			continue
		}

		status[resp.Action.ID] = append(status[resp.Action.ID], state)
	}

	if len(emptyChecker) == len(outputs) {
		return
	}

	// now calculate the state of the outputs
	for _, output := range outputs {
		skip := false
		for i := range emptyChecker {
			if output.Device.ID == emptyChecker[i] {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		deepest := output

		var prevEdge graph.Edge
		var prevState input
		search := traverse.DepthFirst{
			Visit: func(node gonum.Node) {
				deepest = node.(graph.Node)
			},
			Traverse: func(edge gonum.Edge) bool {
				e := edge.(graph.Edge)

				states := status[e.Src.Device.ID]

				if prevEdge != (graph.Edge{}) && *prevState.Video != "" {
					if len(states) == 0 && *prevState.Video == e.Dst.Device.Address {
						prevEdge = e
						return true
					}
				}

				if _, ok := e.Src.Type.Commands["GetInput"]; ok {
					for _, state := range states {
						if *state.Video == "" {
							continue
						}

						inputStr := *state.Video
						split := strings.Split(inputStr, ":")
						if len(split) > 1 {
							inputStr = split[1]
						}

						if e.SrcPort.Name == inputStr {
							prevEdge = e
							prevState = state
							return true
						}

						// well we took off outgoing so idk how this needs to change yet
						// if len(e.Src.Device.Ports.Outgoing()) == 1 {
						// 	prevState = state
						// 	prevEdge = e
						// 	return true
						// }
					}

					return false
				}

				if _, ok := e.Src.Type.Commands["GetInputByOutput"]; ok {
					for _, state := range states {
						if *state.Video == "" {
							continue
						}

						inputStr := *state.Video
						split := strings.Split(inputStr, ":")
						if len(split) > 1 {
							inputStr = split[1]
						}
						if prevEdge == (graph.Edge{}) {
							if inputStr == e.SrcPort.Name {
								prevState = state
								prevEdge = e
								return true
							}
						} else {
							if len(split) > 1 {
								if split[1] == prevEdge.DstPort.Name && e.SrcPort.Name == split[0] {
									prevState = state
									prevEdge = e
									return true
								}
							} else {
								if inputStr == e.SrcPort.Name {
									prevState = state
									prevEdge = e
									return true
								}
							}
						}
					}
					return false
				}

				// well we took off outgoing so idk how this needs to change yet
				// if len(e.Src.Device.Ports.Outgoing()) == 1 {
				// 	prevEdge = e
				// 	return true
				// }

				if *prevState.Video == e.Dst.Address {
					prevEdge = e
					return true
				}

				return false
			},
		}

		search.Walk(t, output, func(node gonum.Node) bool {
			return t.From(node.ID()).Len() == 0
		})

		// validate that the deepest is an input
		valid := false
		for _, input := range inputs {
			if deepest.Device.ID == input.Device.ID {
				valid = true
				break
			}
		}
		if valid {
			i := api.Input{
				Video: &deepest.Device.ID,
			}
			resps[0].Updates <- OutputStateUpdate{
				ID: output.Device.ID,
				OutputState: api.OutputState{
					Input: &i,
				},
			}

		} else {
			states := status[deepest.Device.ID]

			resps[0].Errors <- api.DeviceStateError{
				ID:    output.Device.ID,
				Field: "input",
				Error: fmt.Sprintf("unable to traverse input tree back to a valid input. only got to %s|%+v", deepest.Device.ID, states),
			}
			resps[0].Updates <- OutputStateUpdate{}
		}
	}
}
