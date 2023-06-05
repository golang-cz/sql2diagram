CREATE TABLE orgs
(
    id          BIGSERIAL PRIMARY KEY                                NOT NULL,
    external_id UUID                                                 NOT NULL,
    name        CHARACTER VARYING(512) DEFAULT ''::CHARACTER VARYING NOT NULL
);

CREATE TABLE users
(
    id          BIGSERIAL PRIMARY KEY       NOT NULL,
    org_id      BIGINT                      NOT NULL REFERENCES orgs (id),
    external_id UUID                        NOT NULL,
    email       CHARACTER VARYING(255)      NOT NULL,
    role        SMALLINT                    NOT NULL,
    created_at  TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at  TIMESTAMP WITHOUT TIME ZONE
);

CREATE INDEX users_org_id_idx ON users USING btree (org_id);

CREATE TABLE activities
(
    id          BIGSERIAL PRIMARY KEY NOT NULL,
    org_id      BIGINT                NOT NULL REFERENCES orgs (id),
    user_id     BIGINT                NOT NULL REFERENCES users (id),
    external_id UUID                  NOT NULL,
    type        SMALLINT              NOT NULL
);

CREATE INDEX activities_org_id_idx ON activities USING btree (org_id);
CREATE INDEX activities_user_id_idx ON activities USING btree (user_id);

CREATE TABLE questions
(
    id           BIGSERIAL PRIMARY KEY NOT NULL,
    external_id  UUID                  NOT NULL,
    activity_id  BIGINT                NOT NULL REFERENCES activities (id),
    language     VARCHAR               NOT NULL,
    body         TEXT                  NOT NULL,
    min_duration INTEGER               NOT NULL,
    max_duration INTEGER               NOT NULL,
    sensitive    BOOLEAN               NOT NULL DEFAULT FALSE
);

CREATE INDEX questions_activity_id_idx ON questions USING btree (activity_id);

CREATE TABLE videos
(
    id            BIGSERIAL PRIMARY KEY NOT NULL,
    question_id   BIGINT                NOT NULL REFERENCES questions (id),
    respondent_id UUID                  NOT NULL,
    member_id     UUID                  NOT NULL,
    duration      INT                   NOT NULL,
    file_path     VARCHAR               NOT NULL,
    metadata      JSONB                 NOT NULL,
    status        SMALLINT              NOT NULL,
    thumbnail     VARCHAR               NULL
);

CREATE INDEX videos_question_id_idx ON videos USING btree (question_id);

CREATE TABLE topics
(
    id        BIGSERIAL PRIMARY KEY NOT NULL,
    video_id  BIGINT                NOT NULL REFERENCES videos (id),
    name      VARCHAR               NOT NULL,
    sentiment VARCHAR               NOT NULL
);

CREATE INDEX topics_question_id_idx ON topics USING btree (video_id);

CREATE TABLE tags
(
    id       BIGSERIAL PRIMARY KEY NOT NULL,
    video_id BIGINT                NOT NULL REFERENCES videos (id),
    name     VARCHAR               NOT NULL
);

CREATE INDEX tags_question_id_idx ON tags USING btree (video_id);

CREATE TABLE showreels
(
    id        BIGSERIAL PRIMARY KEY NOT NULL,
    file_path VARCHAR               NOT NULL
);

CREATE TABLE video_showreels
(
    id          BIGSERIAL PRIMARY KEY NOT NULL,
    video_id    BIGINT                NOT NULL REFERENCES videos (id),
    showreel_id BIGINT                NOT NULL REFERENCES showreels (id)
);

CREATE INDEX video_showreels_video_id_idx ON video_showreels USING btree (video_id);
CREATE INDEX video_showreels_showreel_id_idx ON video_showreels USING btree (showreel_id);
