// frontend/src/api.ts

const API_BASE_URL = 'https://atomquest-portal-w7cw.onrender.com';

// Helper to grab the token saved by your auth.ts script
const getToken = () => localStorage.getItem('jwt_token');

export const api = {
    async post(endpoint: string, data: any) {
        const response = await fetch(`${API_BASE_URL}${endpoint}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${getToken()}` // Injects the secure token
            },
            body: JSON.stringify(data)
        });
        return response.json();
    },

    async put(endpoint: string, data: any) {
        const response = await fetch(`${API_BASE_URL}${endpoint}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${getToken()}`
            },
            body: JSON.stringify(data)
        });
        return response.json();
    },

    async get(endpoint: string) {
        const response = await fetch(`${API_BASE_URL}${endpoint}`, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${getToken()}`
            }
        });
        return response.json();
    }
};