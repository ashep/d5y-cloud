# Raion boundaries

`raions.json.gz` contains simplified Ukrainian raion (admin level 2) boundaries and hromada administrative
centers, derived from "Ukraine - Subnational Administrative Boundaries" by OCHA, licensed under
[CC BY-IGO](https://creativecommons.org/licenses/by/3.0/igo/):
https://data.humdata.org/dataset/cod-ab-ukr (version v05, valid on 2025-09-01).

Changes: geometry simplified (Douglas-Peucker, 0.001°), coordinates rounded to 4 decimals,
attributes reduced to Ukrainian names and pcodes.

To regenerate, download `ukr_admin_boundaries.geojson.zip` from the page above, unpack it and run:

```bash
go run ./internal/alertsinua/genraions -in ukr_admin2.geojson -capitals ukr_admincapitals.geojson -out internal/alertsinua/data/raions.json.gz
```
