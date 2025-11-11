# Mapbox Integration

## Setup Instructions

1. **Get a Mapbox Access Token**
   - Go to [Mapbox](https://www.mapbox.com/)
   - Sign up or log in
   - Navigate to your [Account page](https://account.mapbox.com/)
   - Copy your default public token or create a new one

2. **Configure Environment Variables**
   - Create a `.env` file in the project root (or use `.env.local`)
   - Add your Mapbox token:
     ```
     PUBLIC_MAPBOX_TOKEN=your_actual_mapbox_token_here
     ```

3. **Access the Map**
   - Navigate to `/map` route in your browser
   - The map should display with Bangkok as the default center

## Customization

### Change Map Style
Edit the `style` property in `/src/routes/map/+page.svelte`:
```javascript
style: 'mapbox://styles/mapbox/streets-v12'
```

Available styles:
- `streets-v12` - Standard street map
- `outdoors-v12` - Outdoor/terrain map
- `light-v11` - Light theme
- `dark-v11` - Dark theme
- `satellite-v9` - Satellite imagery
- `satellite-streets-v12` - Satellite with streets

### Change Default Location
Edit the `center` coordinates:
```javascript
center: [longitude, latitude]
```

### Add Custom Markers
```javascript
new mapboxgl.Marker()
    .setLngLat([lng, lat])
    .setPopup(new mapboxgl.Popup().setHTML('<h3>Location Name</h3>'))
    .addTo(map);
```
