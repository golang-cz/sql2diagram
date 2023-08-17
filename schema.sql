
SELECT pg_catalog.set_config('search_path', '', false);

CREATE TABLE public.activities (
                                   id uuid NOT NULL,
                                   org_id uuid NOT NULL,
                                   user_id uuid NOT NULL,
                                   name character varying(512) NOT NULL,
                                   type smallint NOT NULL,
                                   created_at timestamp without time zone NOT NULL,
                                   updated_at timestamp without time zone NOT NULL,
                                   deleted_at timestamp without time zone,
                                   min_length_seconds bigint NOT NULL,
                                   max_length_seconds bigint NOT NULL,
                                   enable_audio_only boolean NOT NULL
);

CREATE TABLE public.files (
                              id uuid NOT NULL,
                              org_id uuid,
                              user_id uuid,
                              response_id uuid,
                              path character varying(512) DEFAULT ''::character varying NOT NULL,
                              type smallint NOT NULL,
                              metadata jsonb
);

CREATE TABLE public.goose_db_version (
                                         id integer NOT NULL,
                                         version_id bigint NOT NULL,
                                         is_applied boolean NOT NULL,
                                         tstamp timestamp without time zone DEFAULT now()
);

CREATE SEQUENCE public.goose_db_version_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.goose_db_version_id_seq OWNED BY public.goose_db_version.id;

CREATE TABLE public.ivs_channels (
                                     id uuid NOT NULL,
                                     arn character varying(255) NOT NULL,
                                     ingest_endpoint character varying(512) DEFAULT ''::character varying NOT NULL,
                                     stream_key character varying(512) DEFAULT ''::character varying NOT NULL,
                                     created_at timestamp without time zone NOT NULL,
                                     updated_at timestamp without time zone NOT NULL,
                                     deleted_at timestamp without time zone
);

CREATE TABLE public.ivs_events (
                                   id uuid NOT NULL,
                                   log_stream_name character varying(512),
                                   detail_type smallint NOT NULL,
                                   occurred_at timestamp without time zone NOT NULL,
                                   message jsonb NOT NULL,
                                   created_at timestamp without time zone NOT NULL,
                                   event_name smallint,
                                   reason character varying,
                                   stream_id character varying,
                                   recording_s3_key_prefix character varying
);

CREATE TABLE public.mediaconvert_jobs (
                                          id uuid NOT NULL,
                                          job_id character varying NOT NULL,
                                          status smallint DEFAULT 0 NOT NULL,
                                          warnings jsonb,
                                          error_code integer,
                                          error_message text,
                                          created_at timestamp without time zone,
                                          finished_at timestamp without time zone,
                                          deleted_at timestamp without time zone
);

CREATE TABLE public.orgs (
                             id uuid NOT NULL,
                             external_application_id uuid NOT NULL,
                             name character varying(512) DEFAULT ''::character varying NOT NULL,
                             created_at timestamp without time zone NOT NULL,
                             updated_at timestamp without time zone,
                             deleted_at timestamp without time zone,
                             feature_toggles jsonb
);

CREATE TABLE public.questions (
                                  id uuid NOT NULL,
                                  activity_id uuid NOT NULL,
                                  locale smallint NOT NULL,
                                  body text NOT NULL,
                                  min_duration integer NOT NULL,
                                  max_duration integer NOT NULL,
                                  sensitive boolean DEFAULT false NOT NULL
);

CREATE TABLE public.responses (
                                  id uuid NOT NULL,
                                  respondent_id uuid NOT NULL,
                                  member_id uuid NOT NULL,
                                  question_id uuid NOT NULL,
                                  status smallint NOT NULL,
                                  test boolean DEFAULT false NOT NULL,
                                  duration integer NOT NULL,
                                  browser_locale character varying,
                                  ip_address character varying,
                                  inappropriate boolean,
                                  sensitive boolean,
                                  transcript text,
                                  original_transcript text,
                                  summary character varying,
                                  notes text
);

CREATE TABLE public.showreel_responses (
                                           id uuid NOT NULL,
                                           response_id uuid NOT NULL,
                                           showreel_id uuid NOT NULL
);

CREATE TABLE public.showreels (
    id uuid NOT NULL
);

CREATE TABLE public.tags (
                             id uuid NOT NULL,
                             response_id uuid NOT NULL,
                             name character varying NOT NULL
);

CREATE TABLE public.topics (
                               id uuid NOT NULL,
                               response_id uuid NOT NULL,
                               name character varying NOT NULL,
                               sentiment character varying NOT NULL
);

CREATE TABLE public.transcription_jobs (
                                           id uuid NOT NULL,
                                           name character varying(512),
                                           status smallint,
                                           failed_reason text,
                                           started_at timestamp without time zone NOT NULL,
                                           completed_at timestamp without time zone NOT NULL,
                                           created_at timestamp without time zone NOT NULL,
                                           updated_at timestamp without time zone NOT NULL
);

CREATE TABLE public.users (
                              id uuid NOT NULL,
                              external_user_id uuid NOT NULL,
                              org_id uuid NOT NULL,
                              email character varying(255) NOT NULL,
                              firstname character varying(512) DEFAULT ''::character varying NOT NULL,
                              lastname character varying(512) DEFAULT ''::character varying NOT NULL,
                              role smallint NOT NULL,
                              created_at timestamp without time zone NOT NULL,
                              updated_at timestamp without time zone,
                              deleted_at timestamp without time zone,
                              permissions jsonb
);

ALTER TABLE ONLY public.goose_db_version ALTER COLUMN id SET DEFAULT nextval('public.goose_db_version_id_seq'::regclass);

ALTER TABLE ONLY public.activities
    ADD CONSTRAINT activities_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.files
    ADD CONSTRAINT files_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.goose_db_version
    ADD CONSTRAINT goose_db_version_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.ivs_channels
    ADD CONSTRAINT ivs_channels_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.ivs_events
    ADD CONSTRAINT ivs_events_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.mediaconvert_jobs
    ADD CONSTRAINT mediaconvert_jobs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.orgs
    ADD CONSTRAINT orgs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.questions
    ADD CONSTRAINT questions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.responses
    ADD CONSTRAINT responses_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.showreel_responses
    ADD CONSTRAINT showreel_responses_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.showreels
    ADD CONSTRAINT showreels_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.topics
    ADD CONSTRAINT topics_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.transcription_jobs
    ADD CONSTRAINT transcription_jobs_name_key UNIQUE (name);

ALTER TABLE ONLY public.transcription_jobs
    ADD CONSTRAINT transcription_jobs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

CREATE INDEX activities_org_id_idx ON public.activities USING btree (org_id);

CREATE INDEX activities_user_id_idx ON public.activities USING btree (user_id);

CREATE INDEX questions_activity_id_idx ON public.questions USING btree (activity_id);

CREATE INDEX showreel_responses_response_id_idx ON public.showreel_responses USING btree (response_id);

CREATE INDEX showreel_responses_showreel_id_idx ON public.showreel_responses USING btree (showreel_id);

CREATE INDEX tags_response_id_idx ON public.tags USING btree (response_id);

CREATE INDEX topics_response_id_idx ON public.topics USING btree (response_id);

CREATE INDEX users_org_id_idx ON public.users USING btree (org_id);

CREATE INDEX videos_question_id_idx ON public.responses USING btree (question_id);

ALTER TABLE ONLY public.activities
    ADD CONSTRAINT activities_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.orgs(id);

ALTER TABLE ONLY public.activities
    ADD CONSTRAINT activities_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

ALTER TABLE ONLY public.files
    ADD CONSTRAINT files_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.orgs(id);

ALTER TABLE ONLY public.files
    ADD CONSTRAINT files_response_id_fkey FOREIGN KEY (response_id) REFERENCES public.responses(id);

ALTER TABLE ONLY public.files
    ADD CONSTRAINT files_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

ALTER TABLE ONLY public.questions
    ADD CONSTRAINT questions_activity_id_fkey FOREIGN KEY (activity_id) REFERENCES public.activities(id);

ALTER TABLE ONLY public.responses
    ADD CONSTRAINT responses_question_id_fkey FOREIGN KEY (question_id) REFERENCES public.questions(id);

ALTER TABLE ONLY public.showreel_responses
    ADD CONSTRAINT showreel_responses_response_id_fkey FOREIGN KEY (response_id) REFERENCES public.responses(id);

ALTER TABLE ONLY public.showreel_responses
    ADD CONSTRAINT showreel_responses_showreel_id_fkey FOREIGN KEY (showreel_id) REFERENCES public.showreels(id);

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_response_id_fkey FOREIGN KEY (response_id) REFERENCES public.responses(id);

ALTER TABLE ONLY public.topics
    ADD CONSTRAINT topics_response_id_fkey FOREIGN KEY (response_id) REFERENCES public.responses(id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.orgs(id);

