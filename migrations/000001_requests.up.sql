CREATE TABLE requests (
    id BIGSERIAL PRIMARY KEY,
    user_name VARCHAR(255) NOT NULL,
    mobile_number VARCHAR(20) NOT NULL,
    trip_type VARCHAR(20) NOT NULL CHECK (trip_type IN ('inter', 'intra')),
    origin_lat DOUBLE PRECISION NOT NULL,
    origin_long DOUBLE PRECISION NOT NULL,
    origin_index BIGINT NOT NULL,
    dest_lat DOUBLE PRECISION NOT NULL,
    dest_long DOUBLE PRECISION NOT NULL,
    dest_index BIGINT NOT NULL,
    distance DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
