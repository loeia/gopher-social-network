-- +goose Up

-- +goose StatementBegin
CREATE FUNCTION validate_comment_parent() RETURNS trigger AS $$
BEGIN
    IF NEW.parent_id IS NOT NULL THEN
        IF NOT EXISTS (
            SELECT 1 FROM comments
            WHERE id = NEW.parent_id AND post_id = NEW.post_id
        ) THEN
            RAISE EXCEPTION 'parent comment must belong to the same post';
        END IF;
    END IF;
    RETURN NEW;
END $$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_validate_comment_parent
    BEFORE INSERT OR UPDATE ON comments
    FOR EACH ROW EXECUTE FUNCTION validate_comment_parent();

-- +goose Down
DROP TRIGGER IF EXISTS trg_validate_comment_parent ON comments;
DROP FUNCTION IF EXISTS validate_comment_parent();
