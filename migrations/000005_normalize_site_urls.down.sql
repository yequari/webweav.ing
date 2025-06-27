UPDATE websites SET SiteUrl = substr(SiteUrl, 8) WHERE substr(SiteUrl, 1, 4) = 'http';
