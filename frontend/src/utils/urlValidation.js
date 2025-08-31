export const validateAndNormalizeURL = (input) => {
  if (!input || typeof input !== 'string') {
    return { isValid: false, error: 'Please enter a valid URL' };
  }

  let url = input.trim();

  	if (!url.match(/^https?:\/\//)) {
		url = `https://${url}`;
	}

	try {
		const urlObj = new URL(url);

		if (!urlObj.hostname || urlObj.hostname.length === 0) {
			return { isValid: false, error: 'Please enter a valid domain name' };
		}

		if (!urlObj.hostname.includes('.')) {
			return { isValid: false, error: 'Please enter a valid domain name (e.g., example.com)' };
		}

		const parts = urlObj.hostname.split('.');
    if (parts.length < 2 || parts[parts.length - 1].length < 2) {
      return { isValid: false, error: 'Please enter a valid domain name with proper TLD' };
    }

    return { 
      isValid: true, 
      normalizedUrl: url,
      originalInput: input.trim()
    };
  } catch (error) {
    return { isValid: false, error: 'Please enter a valid URL format' };
  }
};

export const isURL = (string) => {
  try {
    new URL(string);
    return true;
  } catch {
    return false;
  }
};
