-- +goose Up
-- +goose StatementBegin
CREATE TABLE Projects (
                        id serial PRIMARY KEY,
                        key TEXT,
                        title TEXT
);

CREATE TABLE Author (
                        id serial PRIMARY KEY,
                        name TEXT
);

CREATE TABLE Issue (
                       id serial PRIMARY KEY,
                       projectId INT NOT NULL,
                       FOREIGN KEY (projectId) REFERENCES Projects (id) ON DELETE CASCADE ON UPDATE CASCADE,
                       authorId INT NOT NULL,
                       FOREIGN KEY (authorId) REFERENCES Author (id) ON DELETE CASCADE ON UPDATE CASCADE,
                       assigneeId INT NOT NULL,
                       key TEXT,
                       summary TEXT,
                       description TEXT,
                       type TEXT,
                       priority TEXT,
                       status TEXT,
                       createdTime TIMESTAMP WITHOUT TIME ZONE,
                       closedTime TIMESTAMP WITHOUT TIME ZONE,
                       updatedTime TIMESTAMP WITHOUT TIME ZONE,
                       timeSpent INT
);

CREATE TABLE StatusChanges (
                               issueId INT NOT NULL,
                               FOREIGN KEY (issueId) REFERENCES Issue (id) ON DELETE CASCADE ON UPDATE CASCADE,
                               authorId INT NOT NULL,
                               FOREIGN KEY (authorId) REFERENCES Author (id) ON DELETE CASCADE ON UPDATE CASCADE,
                               changeTime TIMESTAMP WITHOUT TIME ZONE,
                               fromStatus TEXT,
                               toStatus TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "StatusChanges";
DROP TABLE IF EXISTS "Issue";
DROP TABLE IF EXISTS "Author";
DROP TABLE IF EXISTS "Projects";
-- +goose StatementEnd
