alter table resources
    add resource_unit int;

alter table resources
    add resource_group text;

alter table resources
    add is_expand boolean;