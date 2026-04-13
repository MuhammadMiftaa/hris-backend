-- +goose Up
-- +goose StatementBegin
INSERT INTO branches (id, code, name, address, latitude, longitude, radius_meters, allow_wfh)
VALUES (
  1, 
  'WAFA', 
  'Wafa Indonesia', 
  'Jl. Raya Wisma Pagesangan No.9 Menanggal, Kec. Gayungan, Kota Surabaya, Jawa Timur 60234', 
  -7.339062308729173, 
  112.72026237301269, 
  100, 
  false
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM branches WHERE id = 1;
-- +goose StatementEnd
