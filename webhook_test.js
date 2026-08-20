const http = require('http');
const fs = require('fs');

const server = http.createServer((req, res) => {
    if (req.method === 'POST' && req.url === '/webhook') {
        let body = '';
        req.on('data', chunk => {
            body += chunk.toString();
        });
        req.on('end', () => {
            let data = {};
            try {
                data = JSON.parse(body);
            } catch (e) {
                data = body;
            }
            const logStr = `\n[${new Date().toISOString()}] Received Webhook:\n${JSON.stringify(data, null, 2)}\n`;
            fs.appendFileSync('webhook.log', logStr);
            console.log(logStr);
            res.writeHead(200, { 'Content-Type': 'text/plain' });
            res.end('OK');
        });
    } else {
        res.writeHead(404);
        res.end();
    }
});

server.listen(9999, () => {
    console.log('Webhook listener on 9999');
});
