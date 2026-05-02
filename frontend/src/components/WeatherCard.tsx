import type { WeatherResponse } from '../api';

const weatherLabels: Record<number, string> = {
  0: 'Clear sky',
  1: 'Mainly clear',
  2: 'Partly cloudy',
  3: 'Overcast',
  45: 'Fog',
  48: 'Depositing rime fog',
  51: 'Light drizzle',
  53: 'Moderate drizzle',
  55: 'Dense drizzle',
  61: 'Slight rain',
  63: 'Moderate rain',
  65: 'Heavy rain',
  71: 'Slight snow',
  73: 'Moderate snow',
  75: 'Heavy snow',
  80: 'Rain showers',
  95: 'Thunderstorm',
};

interface Props {
  weather: WeatherResponse;
}

export default function WeatherCard({ weather }: Props) {
  return (
    <div className="weather-card">
      <div className="weather-card-header">
        <div>
          <h3>{weather.city}</h3>
          <p>{new Date(weather.time).toLocaleString()}</p>
        </div>
        <span className="weather-code">{weatherLabels[weather.weatherCode] ?? 'Stable'}</span>
      </div>
      <div className="weather-metrics">
        <div>
          <strong>{Math.round(weather.temperature)}°C</strong>
          <span>Temperature</span>
        </div>
        <div>
          <strong>{weather.windSpeed.toFixed(1)} km/h</strong>
          <span>Wind speed</span>
        </div>
      </div>
    </div>
  );
}
