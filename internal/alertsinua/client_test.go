package alertsinua

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestLocationUID(t *testing.T) {
	tests := []struct {
		region, city string
		want         int
	}{
		// weatherapi.com transliteration
		{"Kyyivs'ka Oblast'", "Bucha", uidKyivska},
		{"Kyyivs'ka Oblast'", "Kyiv", uidKyiv},
		{"Kharkivs'ka Oblast'", "Kharkiv", uidKharkivska},
		{"L'vivs'ka Oblast'", "Lviv", uidLvivska},
		{"Ivano-Frankivs'ka Oblast'", "Kolomyya", uidIvanoFrankivska},
		{"Khmel'nyts'ka Oblast'", "Kam'yanets'-Podil's'kyy", uidKhmelnytska},
		{"Ternopil's'ka Oblast'", "Ternopil", uidTernopilska},
		{"Zaporiz'ka Oblast'", "Zaporizhzhya", uidZaporizka},
		{"Mykolayivs'ka Oblast'", "Mykolayiv", uidMykolaivska},
		{"Odes'ka Oblast'", "Odesa", uidOdeska},
		{"Zakarpats'ka Oblast'", "Uzhhorod", uidZakarpatska},
		{"Chernivets'ka Oblast'", "Chernivtsi", uidChernivetska},
		{"Chernihivs'ka Oblast'", "Chernihiv", uidChernihivska},
		{"Krym", "Simferopol", uidCrimea},
		{"Sevastopol'", "Sevastopol'", uidSevastopol},

		// English
		{"Kyiv Oblast", "Irpin", uidKyivska},
		{"Kyiv City", "", uidKyiv},
		{"Kyiv", "", uidKyiv},
		{"Dnipropetrovsk Oblast", "Dnipro", uidDnipropetrovska},

		// Ukrainian
		{"Київська область", "Бровари", uidKyivska},
		{"м. Київ", "м. Київ", uidKyiv},
		{"Волинська область", "Луцьк", uidVolynska},
		{"Автономна Республіка Крим", "", uidCrimea},

		{"Mazowieckie", "Warsaw", uidUnknown},
		{"", "", uidUnknown},
	}

	for _, tt := range tests {
		if got := locationUID(tt.region, tt.city); got != tt.want {
			t.Errorf("locationUID(%q, %q) = %d, want %d", tt.region, tt.city, got, tt.want)
		}
	}
}

func TestLoadRaions(t *testing.T) {
	rs, err := loadRaions()
	if err != nil {
		t.Fatal(err)
	}

	if len(rs) != 139 {
		t.Errorf("got %d raions, want 139", len(rs))
	}

	oblasts := make(map[int]bool)
	for _, r := range rs {
		oblasts[r.oblastUID] = true
	}
	if len(oblasts) != 27 {
		t.Errorf("got %d oblasts, want 27", len(oblasts))
	}
}

func TestLocate(t *testing.T) {
	tests := []struct {
		name         string
		lat, lng     float64
		region, city string
		want         Place
	}{
		{"Kyiv center", 50.4501, 30.5234, "Kyyivs'ka Oblast'", "Pushchavoditsa", Place{uidKyiv, "Київ"}},
		{"Kyiv Darnytsia", 50.4547, 30.6206, "Kyyivs'ka Oblast'", "Darnitsa", Place{uidKyiv, "Київ"}},
		{"Bucha", 50.5436, 30.2120, "Kyyivs'ka Oblast'", "Bucha", Place{uidKyivska, "Бучанський"}},
		{"Brovary", 50.5110, 30.7909, "Kyyivs'ka Oblast'", "Brovary", Place{uidKyivska, "Броварський"}},
		{"Dnipro", 48.4647, 35.0462, "Dnipropetrovs'ka Oblast'", "Dnipro", Place{uidDnipropetrovska, "Дніпровський"}},
		{"Uzhhorod", 48.6208, 22.2879, "Zakarpats'ka Oblast'", "Uzhhorod", Place{uidZakarpatska, "Ужгородський"}},
		{"Lviv", 49.8397, 24.0297, "L'vivs'ka Oblast'", "Lviv", Place{uidLvivska, "Львівський"}},
		{"Mykolaiv", 46.9750, 31.9946, "Mykolayivs'ka Oblast'", "Mykolaiv", Place{uidMykolaivska, "Миколаївський"}},
		{"Sevastopol", 44.6166, 33.5254, "Sevastopol'", "Sevastopol", Place{uidSevastopol, "Севастополь"}},
		{"Simferopol", 44.9521, 34.1024, "Krym", "Simferopol", Place{uidCrimea, "Сімферопольський"}},
		{"no coordinates", 0, 0, "Kharkivs'ka Oblast'", "Kharkiv", Place{uidKharkivska, ""}},
		{"outside Ukraine", 52.2297, 21.0122, "Mazowieckie", "Warsaw", Place{uidUnknown, ""}},
	}

	for _, tt := range tests {
		if got := Locate(tt.lat, tt.lng, tt.region, tt.city); got != tt.want {
			t.Errorf("%s: Locate() = %+v, want %+v", tt.name, got, tt.want)
		}
	}
}

func TestIsActive(t *testing.T) {
	alerts := []Alert{
		{ID: 1, Title: "Луганська область", LocationType: "oblast", LocationUID: "16", Oblast: "Луганська область", AlertType: "air_raid"},
		{ID: 2, Title: "м. Київ", LocationType: "city", LocationUID: "31", Oblast: "м. Київ", AlertType: "air_raid"},
		{ID: 3, Title: "Бучанський район", LocationType: "raion", LocationUID: "75", Oblast: "Київська область", AlertType: "air_raid"},
		{ID: 4, Title: "м. Дніпро", LocationType: "city", LocationUID: "332", Oblast: "Дніпропетровська область", AlertType: "air_raid"},
		{ID: 9, Title: "м. Нікополь", LocationType: "city", LocationUID: "350", Oblast: "Дніпропетровська область", AlertType: "air_raid"},
		{ID: 5, Title: "Кам\u2019янський район", LocationType: "raion", LocationUID: "45", Oblast: "Дніпропетровська область", AlertType: "air_raid"},
		{ID: 6, Title: "Покровська територіальна громада", LocationType: "hromada", LocationUID: "351", Oblast: "Дніпропетровська область", Raion: "Синельниківський район", AlertType: "air_raid"},
		{ID: 7, Title: "Запорізька область", LocationType: "oblast", LocationUID: "12", Oblast: "Запорізька область", AlertType: "artillery_shelling"},
		{ID: 8, Title: "Невідомий район", LocationType: "raion", LocationUID: "999", Oblast: "Одеська область", AlertType: "air_raid"},
	}

	st := buildState(alerts, zerolog.Nop())

	tests := []struct {
		place Place
		want  bool
	}{
		{Place{uidLuhanska, "Сіверськодонецький"}, true}, // whole oblast
		{Place{uidKyiv, "Київ"}, true},
		{Place{uidKyivska, "Бучанський"}, true},
		{Place{uidKyivska, "Броварський"}, false},
		{Place{uidKyivska, ""}, true},                          // raion unknown: any raion counts
		{Place{uidDnipropetrovska, "Дніпровський"}, true},      // city alert counts for its raion
		{Place{uidDnipropetrovska, "Кам'янський"}, true},       // apostrophe variants match
		{Place{uidDnipropetrovska, "Синельниківський"}, false}, // hromada alerts are ignored
		{Place{uidDnipropetrovska, "Павлоградський"}, false},
		{Place{uidZaporizka, "Запорізький"}, false}, // not an air raid
		{Place{uidOdeska, "Одеський"}, true},        // unknown raion falls back to the whole oblast
		{Place{uidLvivska, "Львівський"}, false},
		{Place{uidUnknown, ""}, false},
	}

	for _, tt := range tests {
		if got := st.isActive(tt.place); got != tt.want {
			t.Errorf("isActive(%+v) = %v, want %v", tt.place, got, tt.want)
		}
	}

	if len(st.all) != 7 {
		t.Errorf("got %d tracked alerts, want 7", len(st.all))
	}
}
