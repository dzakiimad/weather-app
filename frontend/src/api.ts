import axios from 'axios';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  headers: {
    'Content-Type': 'application/json',
  },
});

export interface Favorite {
  id: number;
  city: string;
  latitude: number;
  longitude: number;
}

export interface WeatherResponse {
  city: string;
  temperature: number;
  windSpeed: number;
  weatherCode: number;
  time: string;
}

export const fetchWeather = async (city: string): Promise<WeatherResponse> => {
  const response = await api.get('/api/weather', {
    params: { city },
  });
  return response.data;
};

export const fetchFavorites = async (): Promise<Favorite[]> => {
  const response = await api.get('/api/favorites');
  return response.data;
};

export const addFavorite = async (city: string): Promise<Favorite> => {
  const response = await api.post('/api/favorites', { city });
  return response.data;
};

export const deleteFavorite = async (id: number): Promise<void> => {
  await api.delete(`/api/favorites/${id}`);
};
