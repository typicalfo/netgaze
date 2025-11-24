# Addendum 7 – Geolocation Collector Details

**Responsibility**: Populate `Report.Geo` with location and ISP information.

**File**: `internal/collector/geo.go`

**Behavior**
- Single function: `collectGeo(ctx context.Context, target string, r *model.Report)`
- Uses ip-api.com free tier (no API key required)
- 4 second timeout
- Queries first resolved IPv4, then IPv6 if needed

**API endpoint**
```
GET http://ip-api.com/json/1.2.3.4?fields=status,message,country,countryCode,region,regionName,city,zip,lat,lon,timezone,isp,org,as,query
```

**Response mapping**
```go
r.Geo.IP = response.Query
r.Geo.City = response.City
r.Geo.Region = response.RegionName
r.Geo.RegionCode = response.Region
r.Geo.Country = response.Country
r.Geo.CountryCode = response.CountryCode
r.Geo.Org = response.Org
r.Geo.ISP = response.ISP
r.Geo.ASN = extractASN(response.AS) // "AS1234 Example Org"
r.Geo.Latitude = response.Lat
r.Geo.Longitude = response.Lon
r.Geo.Timezone = response.Timezone
```

**Error handling**
- HTTP failure → set `Error = "geolocation API request failed"`
- Rate limit → set `Error = "geolocation rate limited"`
- Invalid response → set `Error = "geolocation API error"`
- Add to `r.Errors["geo"]` on complete failure

**Rate limiting**
- Free tier: 45 requests/minute, 1000 requests/day
- Implement simple backoff if rate limited
- Cache responses in memory during single run

**Fallback strategy**
- If ip-api.com fails, try ipinfo.io (also free tier)
- If both fail, leave geo fields empty but continue collection

**Privacy considerations**
- No API keys or authentication required
- Public IP information only
- No user data sent to third parties

**Performance target**
- Complete in ≤ 4 seconds
- Single HTTP request with context timeout

**Data quality**
- Generally accurate for ISP and country level
- City-level accuracy varies by ISP
- Coordinates are approximate (center of city/region)