package params

import (
	"encoding/base64"
	"foxy/internal/utils"
	"strconv"
)

func IntVal(params []string) *int {
	if len(params) == 0 {
		return nil
	}

	i, err := strconv.Atoi(params[0])
	if err != nil {
		return nil
	}

	return &i
}

func FloatVal(params []string) *float64 {
	if len(params) == 0 {
		return nil
	}

	f, err := strconv.ParseFloat(params[0], 64)
	if err != nil {
		return nil
	}

	return &f
}

func NormalizedFloatVal(params []string) *float64 {
	if len(params) == 0 {
		return nil
	}

	f, err := strconv.ParseFloat(params[0], 64)
	if err != nil {
		return nil
	}

	return utils.Ptr(f / 100.0)
}

func BoolVal(params []string) *bool {
	if len(params) == 0 {
		return nil
	}

	b, err := strconv.ParseBool(params[0])
	if err != nil {
		return nil
	}

	return &b
}

func Int2DVectorVal(params []string) (*int, *int) {
	if len(params) == 1 {
		x := IntVal(params)
		return x, x
	} else if len(params) > 1 {
		x := IntVal(params[:1])
		y := IntVal(params[1:])
		return x, y
	}

	return nil, nil
}

func FlexibleInt2DVectorVal(params []string) (*bool, *int, *int) {
	if len(params) < 2 {
		return nil, nil, nil
	}

	relative := utils.Ptr(params[0] == "rel")

	if len(params) == 2 {
		x := IntVal(params[1:])
		return relative, x, x
	} else if len(params) > 2 {
		x := IntVal(params[1:])
		y := IntVal(params[2:])
		return relative, x, y
	}

	return nil, nil, nil
}

func Float2DVectorVal(params []string) (*float64, *float64) {
	if len(params) == 1 {
		x := FloatVal(params)
		return x, x
	} else if len(params) > 1 {
		x := FloatVal(params[:1])
		y := FloatVal(params[1:])
		return x, y
	}

	return nil, nil
}

func FlexibleFloat2DVectorVal(params []string) (*bool, *float64, *float64) {
	if len(params) < 2 {
		return nil, nil, nil
	}

	relative := utils.Ptr(params[0] == "rel")

	if len(params) == 2 {
		x := FloatVal(params[1:])
		return relative, x, x
	} else if len(params) > 2 {
		x := FloatVal(params[1:])
		y := FloatVal(params[2:])
		return relative, x, y
	}

	return nil, nil, nil
}

func IntRectVal(params []string) (*int, *int, *int, *int) {
	if len(params) == 1 {
		x := IntVal(params)
		return x, x, x, x
	} else if len(params) == 2 {
		x := IntVal(params[:1])
		w := IntVal(params[1:])
		return x, x, w, w
	} else if len(params) == 4 {
		x := IntVal(params[:1])
		y := IntVal(params[1:])
		w := IntVal(params[2:])
		h := IntVal(params[3:])
		return x, y, w, h
	}

	return nil, nil, nil, nil
}

func FloatRectVal(params []string) (*float64, *float64, *float64, *float64) {
	if len(params) == 1 {
		x := FloatVal(params)
		return x, x, x, x
	} else if len(params) == 2 {
		x := FloatVal(params[:1])
		w := FloatVal(params[1:])
		return x, x, w, w
	} else if len(params) == 4 {
		x := FloatVal(params[:1])
		y := FloatVal(params[1:])
		w := FloatVal(params[2:])
		h := FloatVal(params[3:])
		return x, y, w, h
	}

	return nil, nil, nil, nil
}

func DecodeBase64StringVal(params []string) *string {
	if len(params) == 0 {
		return nil
	}

	base64str := params[0]
	for len(base64str)%4 != 0 {
		base64str += "="
	}

	str, err := base64.URLEncoding.DecodeString(base64str)
	if err != nil {
		return nil
	}

	return utils.Ptr(string(str))
}
