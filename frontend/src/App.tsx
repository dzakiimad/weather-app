import { useMemo, useState, useEffect } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { addFavorite, deleteFavorite, fetchFavorites, fetchWeather } from './api';
import WeatherCard from './components/WeatherCard';
import CityPicker from './components/CityPicker';

const defaultCity = 'New York';

function App() {
  const [city, setCity] = useState(defaultCity);
  const [darkMode, setDarkMode] = useState(true);

  useEffect(() => {
    document.body.className = darkMode ? 'dark' : 'light';
  }, [darkMode]);

  const queryClient = useQueryClient();
  const weatherQuery = useQuery({
    queryKey: ['weather', city],
    queryFn: () => fetchWeather(city),
    enabled: Boolean(city),
    staleTime: 1000 * 60 * 5,
  });

  const favoritesQuery = useQuery({
    queryKey: ['favorites'],
    queryFn: fetchFavorites,
    staleTime: 1000 * 60 * 10,
  });

  const addFavoriteMutation = useMutation({
    mutationFn: (name: string) => addFavorite(name),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['favorites'] }),
  });

  const deleteFavoriteMutation = useMutation({
    mutationFn: (id: number) => deleteFavorite(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['favorites'] }),
  });

  const favoriteCities = useMemo(
    () => favoritesQuery.data ?? [],
    [favoritesQuery.data]
  );

  const themeClass = useMemo(() => (darkMode ? 'app dark' : 'app light'), [darkMode]);

  return (
    <div className={themeClass}>
      <div className="topbar">
        <div>
          <h1>Weather App</h1>
          <p>Weather insights with offline-friendly favorites.</p>
        </div>
        <button className="toggle-button" onClick={() => setDarkMode((prev) => !prev)}>
          {darkMode ? 'Light Mode' : 'Dark Mode'}
        </button>
      </div>

      <section className="content-grid">
        <div className="panel">
          <CityPicker
            label="Search city"
            value={city}
            onChange={setCity}
            onSubmit={() => weatherQuery.refetch()}
          />

          <button
            className="save-button"
            onClick={() => addFavoriteMutation.mutate(city)}
          >
            Save to favorites
          </button>

          <div className="favorites">
            <h2>Favorite cities</h2>
            {favoritesQuery.isLoading ? (
              <p>Loading favorites…</p>
            ) : (
              <ul>
                {favoriteCities.map((fav) => (
                  <li key={fav.id}>
                    <button className="favorite-link" onClick={() => setCity(fav.city)}>
                      {fav.city}
                    </button>
                    <button
                      className="icon-button"
                      onClick={() => deleteFavoriteMutation.mutate(fav.id)}
                    >
                      ✕
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>

        <div className="panel weather-panel">
          <h2>Current conditions</h2>
          {weatherQuery.isLoading ? (
            <p>Fetching weather for {city}…</p>
          ) : weatherQuery.isError ? (
            <p>Unable to load weather. Try another city.</p>
          ) : weatherQuery.data ? (
            <WeatherCard weather={weatherQuery.data} />
          ) : (
            <p>No weather data yet.</p>
          )}
        </div>
      </section>
    </div>
  );
}

export default App;
