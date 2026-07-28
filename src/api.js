const API = '/api'

async function request(path, options = {}) {
  const res = await fetch(`${API}${path}`, {
    headers: { 'Content-Type': 'application/json', ...options.headers },
    ...options,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || res.statusText)
  }
  if (res.status === 204) return null
  return res.json()
}

export const api = {
  getPersons: () => request('/persons'),
  getShops: () => request('/shops'),
  searchShops: (q) => request(`/shops?q=${encodeURIComponent(q)}`),
  createShop: (name) => request('/shops', { method: 'POST', body: JSON.stringify({ name }) }),
  updateShop: (id, name) =>
    request(`/shops/${id}`, { method: 'PUT', body: JSON.stringify({ name }) }),
  deleteShop: (id) => request(`/shops/${id}`, { method: 'DELETE' }),
  getItems: () => request('/items'),
  searchItems: (q) => request(`/items?q=${encodeURIComponent(q)}`),
  createItem: (name) => request('/items', { method: 'POST', body: JSON.stringify({ name }) }),
  updateItem: (id, name) =>
    request(`/items/${id}`, { method: 'PUT', body: JSON.stringify({ name }) }),
  deleteItem: (id) => request(`/items/${id}`, { method: 'DELETE' }),
  getStats: (params = {}) => {
    const qs = new URLSearchParams()
    if (params.from_date) qs.set('from_date', params.from_date)
    if (params.to_date) qs.set('to_date', params.to_date)
    const query = qs.toString()
    return request(`/stats${query ? `?${query}` : ''}`)
  },
  getExpenses: (params = {}) => {
    const qs = new URLSearchParams()
    if (params.person_id) qs.set('person_id', params.person_id)
    if (params.from_date) qs.set('from_date', params.from_date)
    if (params.to_date) qs.set('to_date', params.to_date)
    const query = qs.toString()
    return request(`/expenses${query ? `?${query}` : ''}`)
  },
  createExpenses: (data) => request('/expenses', { method: 'POST', body: JSON.stringify(data) }),
  updateExpense: (id, data) =>
    request(`/expenses/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteExpense: (id) => request(`/expenses/${id}`, { method: 'DELETE' }),
}
