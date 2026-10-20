DO $$
    begin if exists (
        select 1 from information_schema.columns
        where table_schema ='public'
        and table_name ='products'
        and column_name ='deleted_at'
    ) then 
        alter table products
        drop column deleted_at;
    end if;
end $$;
