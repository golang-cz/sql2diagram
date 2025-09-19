-- Reset search path
select pg_catalog.set_config('search_path', '', false)
;

-- ORGANIZATIONS
CREATE TABLE public.organizations (
    id uuid NOT NULL,
    name varchar(255) NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT now()
);

-- USERS
CREATE TABLE public.users (
    id uuid NOT NULL,
    org_id uuid NOT NULL,
    email varchar(255) NOT NULL,
    full_name varchar(255) NOT NULL,
    role varchar(50) NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT now()
);

-- PROJECTS
CREATE TABLE public.projects (
    id uuid NOT NULL,
    org_id uuid NOT NULL,
    name varchar(255) NOT NULL,
    description text,
    created_at timestamp without time zone NOT NULL DEFAULT now()
);

-- TASKS
CREATE TABLE public.tasks (
    id uuid NOT NULL,
    project_id uuid NOT NULL,
    assignee_id uuid,
    title varchar(255) NOT NULL,
    status varchar(50) NOT NULL DEFAULT 'open',
    due_date date
);

-- COMMENTS
CREATE TABLE public.comments (
    id uuid NOT NULL,
    task_id uuid NOT NULL,
    user_id uuid NOT NULL,
    body text NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT now()
);

-- TAGS
CREATE TABLE public.tags (
    id uuid NOT NULL,
    name varchar(64) NOT NULL
);

-- TASK_TAGS (many-to-many join)
CREATE TABLE public.task_tags (
    id uuid NOT NULL,
    task_id uuid NOT NULL,
    tag_id uuid NOT NULL
);

-- -----------------------------------------------
-- CONSTRAINTS
-- -----------------------------------------------
-- Primary keys
ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.task_tags
    ADD CONSTRAINT task_tags_pkey PRIMARY KEY (id);

-- Foreign keys
ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id);

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_assignee_id_fkey FOREIGN KEY (assignee_id) REFERENCES public.users(id);

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id);

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

ALTER TABLE ONLY public.task_tags
    ADD CONSTRAINT task_tags_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id);

ALTER TABLE ONLY public.task_tags
    ADD CONSTRAINT task_tags_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES public.tags(id);

-- Uniques
ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_name_key UNIQUE (name);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_org_id_name_key UNIQUE (org_id, name);

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_name_key UNIQUE (name);

ALTER TABLE ONLY public.task_tags
    ADD CONSTRAINT task_tags_task_id_tag_id_key UNIQUE (task_id, tag_id);

-- -----------------------------------------------
-- Indexes
-- -----------------------------------------------
CREATE INDEX users_org_id_idx ON public.users (org_id);
CREATE INDEX projects_org_id_idx ON public.projects (org_id);
CREATE INDEX tasks_project_id_idx ON public.tasks (project_id);
CREATE INDEX tasks_assignee_id_idx ON public.tasks (assignee_id);
CREATE INDEX comments_task_id_idx ON public.comments (task_id);
CREATE INDEX comments_user_id_idx ON public.comments (user_id);
CREATE INDEX task_tags_task_id_idx ON public.task_tags (task_id);
