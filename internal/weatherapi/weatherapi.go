package weatherapi

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ashep/d5y/internal/clientinfo"
	"github.com/ashep/d5y/internal/httpcli"
)

var (
	ErrInvalidArgument = errors.New("invalid argument")
)

type ConditionID int

const (
	ConditionUnknown ConditionID = iota
	ConditionClear
	ConditionPartlyCloudy
	ConditionCloudy
	ConditionOvercast
	ConditionMist
	ConditionLightRain
	ConditionMediumRain
	ConditionHeavyRain
	ConditionLightSnow
	ConditionMediumSnow
	ConditionHeavySnow
	ConditionLightSleet
	ConditionHeavySleet
	ConditionThunderstorm
	ConditionFog
	ConditionLightHail
	ConditionHeavyHail
)

type Service struct {
	c      *httpcli.Client
	apiKey string
}

type Location struct {
	Name     string
	Country  string
	Region   string
	Lat      float64
	Lng      float64
	Timezone string
	Time     string
}

type ConditionItem struct {
	Id        ConditionID `json:"id"`
	Title     string      `json:"title"`
	IsDay     int         `json:"is_day"`
	Temp      float64     `json:"temp"`
	FeelsLike float64     `json:"feels_like"`
}

type Data struct {
	Location Location      `json:"location"`
	Current  ConditionItem `json:"current"`
}

type wAPIRespLocation struct {
	Name           string  `json:"name"`
	Country        string  `json:"country"`
	Region         string  `json:"region"`
	Lat            float64 `json:"lat"`
	Lon            float64 `json:"lon"`
	TZID           string  `json:"tz_id"`
	LocalTimeEpoch int     `json:"localtime_epoch"`
	LocalTime      string  `json:"localtime"`
}

type wAPIRespCondition struct {
	Code int    `json:"code"`
	Text string `json:"text"`
	Icon string `json:"icon"`
}

type wAPIRespCurrent struct {
	Temp      float64           `json:"temp_c"`
	FeelsLike float64           `json:"feelslike_c"`
	Pressure  float64           `json:"pressure_mb"`
	Humidity  float64           `json:"humidity"`
	Condition wAPIRespCondition `json:"condition"`
	IsDay     int               `json:"is_day"`
}

type wAPIResp struct {
	Location wAPIRespLocation `json:"location"`
	Current  wAPIRespCurrent  `json:"current"`
}

func New(apiKey string) *Service {
	return &Service{
		c:      httpcli.New(),
		apiKey: apiKey,
	}
}

func (s *Service) GetFromRequest(req *http.Request) (*Data, error) {
	lat, lng := req.URL.Query().Get("lat"), req.URL.Query().Get("lng")

	if lat != "" && lng != "" && !strings.HasPrefix(lat, "0.") && !strings.HasPrefix(lng, "0.") {
		latF, err := strconv.ParseFloat(lat, 32)
		if err != nil {
			return nil, fmt.Errorf("%w: lat: %w", ErrInvalidArgument, err)
		}
		if latF < -90 || latF > 90 {
			return nil, fmt.Errorf("%w: lat is out of range: %v", ErrInvalidArgument, latF)
		}

		lngF, err := strconv.ParseFloat(lng, 32)
		if err != nil {
			return nil, fmt.Errorf("%w: lng: %w", ErrInvalidArgument, err)
		}
		if lngF < -180 || lngF > 180 {
			return nil, fmt.Errorf("%w: lng is out of range: %v", ErrInvalidArgument, latF)
		}

		return s.GetForLocation(latF, lngF)
	}

	ci := clientinfo.FromCtx(req.Context())
	if ci.RemoteAddr == "" {
		return nil, errors.New("missing remote address in client info")
	}

	return s.GetForIPAddr(ci.RemoteAddr)
}

func (s *Service) GetForIPAddr(addr string) (*Data, error) {
	return s.get(addr)
}

func (s *Service) GetForLocation(lat, lng float64) (*Data, error) {
	return s.get(fmt.Sprintf("%f,%f", lat, lng))
}

func (s *Service) get(query string) (*Data, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("empty weather api key")
	}

	apiURL := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", s.apiKey, query)
	owRes := &wAPIResp{}

	err := s.c.GetJSON(apiURL, owRes)
	if err != nil {
		return nil, err
	}

	res := &Data{
		Location: Location{
			Name:     owRes.Location.Name,
			Country:  owRes.Location.Country,
			Region:   owRes.Location.Region,
			Lat:      owRes.Location.Lat,
			Lng:      owRes.Location.Lon,
			Timezone: owRes.Location.TZID,
			Time:     owRes.Location.LocalTime,
		},
		Current: ConditionItem{
			Id:        mapWeatherAPIConditionID(owRes.Current.Condition.Code),
			Title:     owRes.Current.Condition.Text,
			IsDay:     owRes.Current.IsDay,
			Temp:      owRes.Current.Temp,
			FeelsLike: owRes.Current.FeelsLike,
		},
	}

	return res, nil
}

func mapWeatherAPIConditionID(id int) ConditionID {
	// https://www.weatherapi.com/docs/weather_conditions.json
	switch id {
	case 1000: // Sunny / clear
		return ConditionClear
	case 1003, // Partly cloudy
		1063: // Patchy rain possible
		return ConditionPartlyCloudy
	case 1006: // Cloudy
		return ConditionCloudy
	case 1009: // Overcast
		return ConditionOvercast
	case 1030: // Mist
		return ConditionMist
	case 1072, // Patchy freezing drizzle possible
		1150, // Patchy light drizzle
		1153, // Light drizzle
		1168, // Freezing drizzle
		1180, // Patchy light rain
		1183, // Light rain
		1198, // Light freezing rain
		1240, // Light rain shower
		1273: // Patchy light rain with thunder
		return ConditionLightRain
	case 1172, // Heavy freezing drizzle
		1186, // Moderate rain at times
		1189, // Moderate rain
		1201, // Moderate or heavy freezing rain
		1243: // Moderate or heavy rain shower

		return ConditionMediumRain
	case 1192, // Heavy rain at times
		1195, // Heavy rain
		1246, // Torrential rain shower
		1276: // Moderate or heavy rain with thunder
		return ConditionHeavyRain
	case 1069, // Patchy sleet possible
		1204, // Light sleet
		1249: // Light sleet showers
		return ConditionLightSleet
	case 1207, // Moderate or heavy sleet
		1252: // Moderate or heavy sleet showers
		return ConditionHeavySleet
	case 1087: // Thundery outbreaks possible
		return ConditionThunderstorm
	case 1066, // Patchy snow possible
		1210, // Patchy light snow
		1213, // Light snow
		1255, // Light snow showers
		1279: // Patchy light snow with thunder
		return ConditionLightSnow
	case 1114, // Blowing snow
		1216, // Patchy moderate snow
		1219, // Moderate snow
		1258: // Moderate or heavy snow showers
		return ConditionMediumSnow
	case 1117, // Blizzard
		1222, // Patchy heavy snow
		1225, // Heavy snow
		1282: // Moderate or heavy snow with thunder
		return ConditionHeavySnow
	case 1135, // Fog
		1147: // Freezing fog
		return ConditionFog
	case 1237, // Ice pellets
		1261: // Light showers of ice pellets
		return ConditionLightHail
	case 1264: // Moderate or heavy showers of ice pellets
		return ConditionHeavyHail
	default:
		return ConditionUnknown
	}
}
