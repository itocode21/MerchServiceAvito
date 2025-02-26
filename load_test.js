import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';

export const options = {
    vus: 200,
    duration: '60s',
    rps: 1000,
    thresholds: {
        http_req_duration: ['p(95)<50'],
        'http_req_failed{expected_error:false}': ['rate<0.0001'], // Только неожиданные ошибки
    },
    // Игнорируем 400 как "неудачу"
    http: {
        discardResponseBodies: true, // Оптимизация
    },
};

const users = new SharedArray('users', function () {
    const result = [];
    for (let i = 1; i <= 200; i++) {
        result.push({
            username: `user${i}`,
            password: '12345',
        });
    }
    return result;
});

const items = ['t-shirt', 'cup', 'book', 'pen', 'socks'];

export function setup() {
    const resetRes = http.get('http://localhost:8080/api/reset');
    check(resetRes, { 'reset success': (r) => r.status === 200 });

    const tokens = [];
    for (let i = 0; i < users.length; i++) {
        const user = users[i];
        const registerRes = http.post('http://localhost:8080/api/register', JSON.stringify({
            username: user.username,
            password: user.password,
        }), { headers: { 'Content-Type': 'application/json' } });
        if (!check(registerRes, { 'register success': (r) => r.status === 200 })) {
            console.log(`Register failed for ${user.username}: status=${registerRes.status}, body=${registerRes.body}`);
            continue;
        }

        const authRes = http.post('http://localhost:8080/api/auth', JSON.stringify({
            username: user.username,
            password: user.password,
        }), { headers: { 'Content-Type': 'application/json' } });
        if (!check(authRes, { 'auth success': (r) => r.status === 200 })) {
            console.log(`Auth failed for ${user.username}: status=${authRes.status}, body=${authRes.body}`);
            continue;
        }

        const token = authRes.json().token;
        if (!token) {
            console.log(`No token received for ${user.username}`);
            continue;
        }

        const infoRes = http.get('http://localhost:8080/api/info', {
            headers: { 'Authorization': `Bearer ${token}` },
        });
        if (!check(infoRes, { 'info check': (r) => r.status === 200 })) {
            console.log(`User ${user.username} not found after registration: status=${infoRes.status}, body=${infoRes.body}`);
            continue;
        }

        tokens.push({ username: user.username, token: token });
    }
    console.log(`Setup completed with ${tokens.length} users registered and verified`);
    return { tokens };
}

export default function (data) {
    const userData = data.tokens[__VU - 1];
    const token = userData.token;
    const username = userData.username;
    const headers = { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' };

    if (Math.random() < 0.8) {
        const infoRes = http.get('http://localhost:8080/api/info', { headers });
        check(infoRes, { 'info success': (r) => r.status === 200 });
    }
    else if (Math.random() < 0.9) {
        const item = items[Math.floor(Math.random() * items.length)];
        const buyRes = http.get(`http://localhost:8080/api/buy/${item}`, { headers, tags: { expected_error: 'true' } });
        check(buyRes, {
            'buy success': (r) => r.status === 200 || r.json().error.includes("недостаточно монет")
        });
    }
    else if (Math.random() < 0.97) {
        const toUser = data.tokens[Math.floor(Math.random() * data.tokens.length)].username;
        const sendRes = http.post('http://localhost:8080/api/sendCoin', JSON.stringify({
            toUser: toUser,
            amount: 10,
        }), { headers, tags: { expected_error: 'true' } });
        check(sendRes, {
            'send success': (r) => r.status === 200 || r.json().error.includes("недостаточно монет")
        });
    }
    else {
        const authRes = http.post('http://localhost:8080/api/auth', JSON.stringify({
            username: username,
            password: '12345',
        }), { headers });
        check(authRes, { 'auth success': (r) => r.status === 200 });
    }

    sleep(0.05);
}