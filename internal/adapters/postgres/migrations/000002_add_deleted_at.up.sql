DO $$
    begin if not exists (
        select 1 from information_schema.columns
        where table_schema ='public'
        and table_name ='products'
        and column_name ='deleted_at'
    ) then 
        alter table products
        add column deleted_at timestamptz;
    end if;
end $$;
