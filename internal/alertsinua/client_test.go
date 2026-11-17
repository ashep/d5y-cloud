package alertsinua

import "testing"

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

func TestAirRaidAlertsByUID(t *testing.T) {
	alerts := []Alert{
		{Title: "Харківська область", LocationType: "oblast", LocationUID: "22", AlertType: "air_raid"},
		{Title: "м. Київ", LocationType: "city", LocationUID: "31", AlertType: "air_raid"},
		{Title: "Нікопольська громада", LocationType: "hromada", LocationUID: "351", AlertType: "air_raid"},
		{Title: "Херсонський район", LocationType: "raion", LocationUID: "125", AlertType: "air_raid"},
		{Title: "м. Харків", LocationType: "city", LocationUID: "124", AlertType: "air_raid"},
		{Title: "Запорізька область", LocationType: "oblast", LocationUID: "12", AlertType: "artillery_shelling"},
		{Title: "Broken", LocationType: "oblast", LocationUID: "", AlertType: "air_raid"},
	}

	got := airRaidAlertsByUID(alerts)

	if len(got) != 2 {
		t.Fatalf("got %d alerts, want 2: %v", len(got), got)
	}
	for _, uid := range []int{uidKharkivska, uidKyiv} {
		if _, ok := got[uid]; !ok {
			t.Errorf("alert for uid %d is missing", uid)
		}
	}
}
