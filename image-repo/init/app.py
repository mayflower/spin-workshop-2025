from spin_sdk import sqlite
from spin_sdk.http import IncomingHandler as IncomingHandlerBase, Request, Response


class IncomingHandler(IncomingHandlerBase):
    def handle_request(self, request: Request) -> Response:
        with sqlite.open_default() as db:
            db.execute("PRAGMA foreign_keys = ON;", [])

            db.execute("""
                CREATE TABLE IF NOT EXISTS originals (
                    name TEXT NOT NULL UNIQUE,
                    created INTEGER NOT NULL,
                    data BLOB
                );
            """, [])

            db.execute("""
                CREATE TABLE IF NOT EXISTS derived (
                    original_name TEXT NOT NULL,
                    transformation TEXT NOT NULL,
                    created INTEGER NOT NULL,
                    last_access INTEGER NOT NULL,
                    data BLOB,
                       
                    CONSTRAINT fk_original_name
                        FOREIGN KEY (original_name)
                        REFERENCES originals(name)
                        ON DELETE CASCADE,
                        
                    CONSTRAINT uk_original_transformation
                        UNIQUE (original_name, transformation)
                );
            """, [])

        return Response(
            200,
            {"content-type": "text/plain"},
            bytes("init complete", "utf-8")
        )
