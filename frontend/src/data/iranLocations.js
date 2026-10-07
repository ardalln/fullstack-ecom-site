import rows from './iranLocations.json';

export const IRAN_PROVINCES = rows.map((province) => province.name);
export const IRAN_LOCATIONS = Object.fromEntries(rows.map((province) => [province.name, {
  cities: province.cities,
  counties: province.counties,
}]));
