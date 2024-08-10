This uses a script to allow you to update your Cloudflare domain DNS using a home dynamic ip address that checks every hour to see if your ip address has been changed.

The script runs to be lightweight and runs every hour on cron.

Set the following variables in your environment

DOMAIN=domain.com;zone_id;dns_record_id|domain2.com;zone_id;dns_record_id
API_KEY=
EMAIL_KEY=

Container image can be found at https://hub.docker.com/r/jbronson29/
