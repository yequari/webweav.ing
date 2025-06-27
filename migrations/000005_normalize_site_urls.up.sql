UPDATE websites SET SiteUrl = 'http://' || SiteUrl WHERE substr(SiteUrl, 1, 4) <> 'http';
