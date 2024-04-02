drop table if exists song;
create table song
(
    id     integer   not null primary key,
    artist char(128) not null,
    title  char(128) not null
);

drop table if exists playmoment;
create table playmoment
(
    id        integer not null primary key,
    timestamp timestamp default current_timestamp,
    songid    integer not null,
    foreign key (songid) references song (id) on delete cascade
);
