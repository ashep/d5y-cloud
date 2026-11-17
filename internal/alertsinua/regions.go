package alertsinua

import (
	"strings"
	"unicode"
)

// Location UIDs as documented at https://devs.alerts.in.ua/#modeluid.
const (
	uidUnknown         = 0
	uidKhmelnytska     = 3
	uidVinnytska       = 4
	uidRivnenska       = 5
	uidVolynska        = 8
	uidDnipropetrovska = 9
	uidZhytomyrska     = 10
	uidZakarpatska     = 11
	uidZaporizka       = 12
	uidIvanoFrankivska = 13
	uidKyivska         = 14
	uidKirovohradska   = 15
	uidLuhanska        = 16
	uidMykolaivska     = 17
	uidOdeska          = 18
	uidPoltavska       = 19
	uidSumska          = 20
	uidTernopilska     = 21
	uidKharkivska      = 22
	uidKhersonska      = 23
	uidCherkaska       = 24
	uidChernihivska    = 25
	uidChernivetska    = 26
	uidLvivska         = 27
	uidDonetska        = 28
	uidCrimea          = 29
	uidSevastopol      = 30
	uidKyiv            = 31
)

// cityAliases maps normalized names of cities that have their own alert location to UIDs.
var cityAliases = map[string]int{
	"kyiv": uidKyiv, "kyyiv": uidKyiv, "kiev": uidKyiv, "київ": uidKyiv,
	"sevastopol": uidSevastopol, "севастополь": uidSevastopol,
}

// regionAliases maps normalized region names (with "oblast" suffix stripped) to UIDs.
// Both Ukrainian and Latin (weatherapi.com transliteration and English) spellings are covered.
var regionAliases = map[string]int{
	"khmelnytska": uidKhmelnytska, "khmelnytskyi": uidKhmelnytska, "khmelnytskyy": uidKhmelnytska,
	"khmelnytsky": uidKhmelnytska, "хмельницька": uidKhmelnytska,

	"vinnytska": uidVinnytska, "vinnytsia": uidVinnytska, "vinnytsya": uidVinnytska, "vinnitsa": uidVinnytska,
	"вінницька": uidVinnytska,

	"rivnenska": uidRivnenska, "rivne": uidRivnenska, "rovno": uidRivnenska, "рівненська": uidRivnenska,

	"volynska": uidVolynska, "volyn": uidVolynska, "волинська": uidVolynska,

	"dnipropetrovska": uidDnipropetrovska, "dnipropetrovsk": uidDnipropetrovska, "dnipro": uidDnipropetrovska,
	"дніпропетровська": uidDnipropetrovska,

	"zhytomyrska": uidZhytomyrska, "zhytomyr": uidZhytomyrska, "житомирська": uidZhytomyrska,

	"zakarpatska": uidZakarpatska, "zakarpattia": uidZakarpatska, "zakarpattya": uidZakarpatska,
	"transcarpathia": uidZakarpatska, "transcarpathian": uidZakarpatska, "закарпатська": uidZakarpatska,

	"zaporizka": uidZaporizka, "zaporizhzhia": uidZaporizka, "zaporizhzhya": uidZaporizka,
	"zaporizhia": uidZaporizka, "zaporozhye": uidZaporizka, "запорізька": uidZaporizka,

	"ivanofrankivska": uidIvanoFrankivska, "ivanofrankivsk": uidIvanoFrankivska,
	"іванофранківська": uidIvanoFrankivska,

	"kyivska": uidKyivska, "kyyivska": uidKyivska, "kievska": uidKyivska, "київська": uidKyivska,

	"kirovohradska": uidKirovohradska, "kirovohrad": uidKirovohradska, "kirovograd": uidKirovohradska,
	"kropyvnytskyi": uidKirovohradska, "кіровоградська": uidKirovohradska,

	"luhanska": uidLuhanska, "luhansk": uidLuhanska, "lugansk": uidLuhanska, "луганська": uidLuhanska,

	"mykolaivska": uidMykolaivska, "mykolayivska": uidMykolaivska, "mykolaiv": uidMykolaivska,
	"mykolayiv": uidMykolaivska, "nikolaev": uidMykolaivska, "миколаївська": uidMykolaivska,

	"odeska": uidOdeska, "odesa": uidOdeska, "odessa": uidOdeska, "одеська": uidOdeska,

	"poltavska": uidPoltavska, "poltava": uidPoltavska, "полтавська": uidPoltavska,

	"sumska": uidSumska, "sumy": uidSumska, "сумська": uidSumska,

	"ternopilska": uidTernopilska, "ternopil": uidTernopilska, "тернопільська": uidTernopilska,

	"kharkivska": uidKharkivska, "kharkiv": uidKharkivska, "kharkov": uidKharkivska, "харківська": uidKharkivska,

	"khersonska": uidKhersonska, "kherson": uidKhersonska, "херсонська": uidKhersonska,

	"cherkaska": uidCherkaska, "cherkasy": uidCherkaska, "cherkasska": uidCherkaska, "черкаська": uidCherkaska,

	"chernihivska": uidChernihivska, "chernihiv": uidChernihivska, "chernigov": uidChernihivska,
	"чернігівська": uidChernihivska,

	"chernivetska": uidChernivetska, "chernivtsi": uidChernivetska, "чернівецька": uidChernivetska,

	"lvivska": uidLvivska, "lviv": uidLvivska, "lvov": uidLvivska, "львівська": uidLvivska,

	"donetska": uidDonetska, "donetsk": uidDonetska, "донецька": uidDonetska,

	"krym": uidCrimea, "crimea": uidCrimea, "avtonomnarespublikakrym": uidCrimea,
	"autonomousrepublicofcrimea": uidCrimea, "respublikakrym": uidCrimea,
	"крим": uidCrimea, "автономнареспублікакрим": uidCrimea,

	"sevastopol": uidSevastopol, "mistosevastopol": uidSevastopol, "sevastopolcity": uidSevastopol,
	"cityofsevastopol": uidSevastopol, "мсевастополь": uidSevastopol, "севастополь": uidSevastopol,
	"містосевастополь": uidSevastopol,

	"kyivcity": uidKyiv, "cityofkyiv": uidKyiv, "mistokyyiv": uidKyiv, "mistokyiv": uidKyiv,
	"мкиїв": uidKyiv, "містокиїв": uidKyiv,
}

// kyivBaseNames are region names which mean the Kyiv oblast when followed by "oblast", and the city otherwise.
var kyivBaseNames = map[string]bool{"kyiv": true, "kyyiv": true, "kiev": true, "київ": true}

var oblastSuffixes = []string{"oblast", "область", "obl", "region"}

// normalize lowercases s and drops everything except letters,
// so "Ivano-Frankivs'ka Oblast'" becomes "ivanofrankivskaoblast".
func normalize(s string) string {
	b := strings.Builder{}
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) {
			b.WriteRune(r)
		}
	}

	return b.String()
}

// locationUID returns the alerts.in.ua UID of the oblast or the special-status city
// for a location described by region and city names.
func locationUID(region, city string) int {
	if uid, ok := cityAliases[normalize(city)]; ok {
		return uid
	}

	r := normalize(region)
	hasOblastSuffix := false
	for _, sfx := range oblastSuffixes {
		if strings.HasSuffix(r, sfx) && len(r) > len(sfx) {
			r = strings.TrimSuffix(r, sfx)
			hasOblastSuffix = true
			break
		}
	}

	if kyivBaseNames[r] {
		if hasOblastSuffix {
			return uidKyivska
		}
		return uidKyiv
	}

	if uid, ok := regionAliases[r]; ok {
		return uid
	}

	return uidUnknown
}
