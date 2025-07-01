import { openDefault, SqliteConnection, ValueBlob } from '@spinframework/spin-sqlite';
import { AutoRouter } from 'itty-router';

interface Metadata {
    name: string;
    created: number;
    length: number;
}

let router = AutoRouter();

function responseOK(data: unknown): Response {
    return new Response(JSON.stringify(data), {
        headers: {
            'Content-Type': 'application/json',
        },
    });
}

function responseError(reason: string, code: number) {
    return new Response(JSON.stringify({ reason }), { status: code });
}

function getDb(): SqliteConnection {
    const connection = openDefault();

    connection.execute('PRAGMA foreign_keys = ON;', []);

    return openDefault();
}

// Happy happy, joy joy, the SDK does not handle Uint8Array correctly, so we have to wrap it ourselves
function sqliteBlob(val: Uint8Array): ValueBlob {
    return { tag: 'blob', val };
}

async function handlePost(body: ReadableStream<Uint8Array> | null, headers: Headers, name: string): Promise<Response> {
    if (!body) return responseError('body missing', 400);

    if (name === 'meta') return responseError('invalid name: "meta" is reserved', 400);

    if (!/^[a-zA-Z0-9._-]+$/.test(name)) {
        return responseError('invalid name: only letters, digits, and . - _ are allowed', 400);
    }

    const transformerResponse = await fetch('http://self.alt/transform', {
        method: 'POST',
        body,
        headers,
    });

    if (!transformerResponse.ok) return transformerResponse;

    try {
        const created = Math.round(Date.now() / 1000);
        const data = new Uint8Array(await transformerResponse.arrayBuffer());

        try {
            getDb().execute('INSERT INTO originals VALUES(?1, ?2, ?3)', [name, created, sqliteBlob(data)]);
        } catch (e: any) {
            const err = e as { payload?: { val?: string } };

            if (err?.payload?.val?.indexOf?.('UNIQUE') === 0) return responseError('duplicate', 409);

            throw e;
        }

        return responseOK({ name, created, length: data.length } as Metadata);
    } catch (e: any) {
        console.log(e, e.payload);
        throw e;
    }
}

async function handleGet(name: string): Promise<Response> {
    const result = getDb().execute('SELECT data FROM originals WHERE name = ?1', [name]);

    const rows = result.rows;
    if (!rows || rows.length === 0) {
        return new Response('Not found', { status: 404 });
    }

    const imageData = rows[0].data as Uint8Array;

    return new Response(imageData, {
        headers: {
            'Content-Type': 'image/png',
        },
    });
}

async function handleDelete(name: string): Promise<Response> {
    const db = getDb();

    const checkResult = db.execute('SELECT 1 FROM originals WHERE name = ?1', [name]);

    if (!checkResult.rows || checkResult.rows.length === 0) {
        return responseError('Not found', 404);
    }

    getDb().execute('DELETE FROM originals WHERE name = ?1', [name]);

    return responseOK({});
}

async function handleGetAllMeta(): Promise<Response> {
    const result = getDb().execute('SELECT name, created, length(data) AS len FROM originals', []);

    const rows = result.rows;

    const originals: Metadata[] = (rows ?? []).map((row) => ({
        name: row.name as string,
        created: Number(row.created as BigInt),
        length: Number(row.len as BigInt),
    }));

    return responseOK({ originals });
}

async function handleGetMeta(name: string): Promise<Response> {
    const result = getDb().execute('SELECT name, created, length(data) AS len FROM originals WHERE name = ?1', [name]);

    const rows = result.rows;
    if (!rows || rows.length === 0) {
        return responseError('Not found', 404);
    }

    const metadata: Metadata = {
        name: rows[0].name as string,
        created: Number(rows[0].created as BigInt),
        length: Number(rows[0].len as BigInt),
    };

    return responseOK(metadata);
}

router
    .get('/originals', () => handleGetAllMeta())
    .get('/originals/meta', () => handleGetAllMeta())
    .get('/originals/meta/:name', ({ name }) => handleGetMeta(name))
    .post('/originals/:name', ({ name, body, headers }) => handlePost(body, headers, name))
    .get('/originals/:name', ({ name }) => handleGet(name))
    .delete('/originals/:name', ({ name }) => handleDelete(name));

//@ts-ignore
addEventListener('fetch', (event: FetchEvent) => {
    event.respondWith(router.fetch(event.request));
});
