CREATE TYPE job_category AS ENUM (
    'engineering_product', 'design_creative', 'sales_marketing',
    'data_analytics', 'operations_support'
);
ALTER TABLE jobs ALTER COLUMN category TYPE job_category USING category::job_category;
