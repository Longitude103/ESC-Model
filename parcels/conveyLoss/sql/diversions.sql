select canal_id, make_timestamp(cast(extract(YEAR from div_dt) as int), cast(extract(MONTH from div_dt) as int), 1, 0, 0, 0) as div_dt, sum(div_amnt_cfs) as div_amnt_cfs
from (select cdj.canal_id, div_dt, sum(div_amnt_cfs) as div_amnt_cfs
      from sw.dailydiversions inner join sw.canal_diversion_jct cdj on dailydiversions.ndnr_id = cdj.ndnr_id
      WHERE div_dt >= '%d-01-01' AND div_dt <= '%d-12-31' group by cdj.canal_id, div_dt) as daily_query
group by canal_id, extract(MONTH from div_dt), extract(YEAR from div_dt) order by canal_id, div_dt;

select canal_id, cast(extract(MONTH from div_dt) as int) as mnth, avg(div_amnt_cfs) as div_avg
from (select canal_id, make_timestamp(cast(extract(YEAR from div_dt) as int), cast(extract(MONTH from div_dt) as int), 1, 0, 0, 0) as div_dt, sum(div_amnt_cfs) as div_amnt_cfs
      from (select cdj.canal_id, div_dt, sum(div_amnt_cfs) as div_amnt_cfs
            from sw.dailydiversions
                inner join sw.canal_diversion_jct cdj on dailydiversions.ndnr_id = cdj.ndnr_id
            WHERE div_dt >= '%d-01-01' AND div_dt <= '%d-12-31' group by cdj.canal_id, div_dt) as daily_query
      group by canal_id, extract(MONTH from div_dt), extract(YEAR from div_dt) order by canal_id, div_dt) as mnth_query
group by canal_id, extract(MONTH from div_dt);

select cdj.canal_id, div_dt, sum(div_amnt_cfs) as div_amnt_cfs
from sw.dailydiversions
    inner join sw.canal_diversion_jct cdj on dailydiversions.ndnr_id = cdj.ndnr_id
WHERE div_dt between '%d-01-01' and '%d-12-31'
group by cdj.canal_id, div_dt order by div_dt;

select canal_id, st_date, end_date, loss_percent
from sw.excess_flow_periods
where st_date between '%d-01-01' and '%d-12-31' and end_date between '%d-01-01' and '%d-12-31';

select canal_id, make_timestamp(cast(extract(YEAR from div_dt) as int), cast(extract(MONTH from div_dt) as int), 1, 0, 0, 0) as div_dt, sum(div_amnt_cfs) as div_amnt_cfs
from (select cdj.canal_id, div_dt, sum(div_amnt_cfs) as div_amnt_cfs
      from sw.dailydiversions
          inner join sw.canal_diversion_jct cdj on dailydiversions.ndnr_id = cdj.ndnr_id
      WHERE div_dt >= '%d-01-01' AND div_dt <= '%d-12-31' %s
      group by cdj.canal_id, div_dt) as daily_query
group by canal_id, extract(MONTH from div_dt), extract(YEAR from div_dt) order by canal_id, div_dt;