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
		{Title: "Ізюмський район", LocationType: "raion", LocationUID: "125", Oblast: "Харківська область", AlertType: "air_raid"},
		{Title: "Харківська область", LocationType: "oblast", LocationUID: "22", Oblast: "Харківська область", AlertType: "air_raid"},
		{Title: "м. Київ", LocationType: "city", LocationUID: "31", Oblast: "м. Київ", AlertType: "air_raid"},
		{Title: "Бучанський район", LocationType: "raion", LocationUID: "75", Oblast: "Київська область", AlertType: "air_raid"},
		{Title: "м. Дніпро", LocationType: "city", LocationUID: "332", Oblast: "Дніпропетровська область", AlertType: "air_raid"},
		{Title: "Покровська громада", LocationType: "hromada", LocationUID: "351", Oblast: "Донецька область", AlertType: "air_raid"},
		{Title: "Запорізька область", LocationType: "oblast", LocationUID: "12", Oblast: "Запорізька область", AlertType: "artillery_shelling"},
		{Title: "Миколаївський район", LocationType: "raion", LocationUID: "98", Oblast: "Миколаївська область", AlertType: "urban_fights"},
		{Title: "Broken", LocationType: "oblast", LocationUID: "", AlertType: "air_raid"},
		{Title: "Unknown", LocationType: "raion", LocationUID: "999", Oblast: "Nowhere", AlertType: "air_raid"},
	}

	got := airRaidAlertsByUID(alerts)

	want := []int{uidKharkivska, uidKyiv, uidKyivska, uidDnipropetrovska}
	if len(got) != len(want) {
		t.Fatalf("got %d alerts, want %d: %v", len(got), len(want), got)
	}
	for _, uid := range want {
		if _, ok := got[uid]; !ok {
			t.Errorf("alert for uid %d is missing", uid)
		}
	}

	if got[uidKharkivska].LocationType != "oblast" {
		t.Errorf("whole-oblast alert should be preferred, got %+v", got[uidKharkivska])
	}
}
