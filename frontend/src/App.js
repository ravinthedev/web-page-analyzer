import React, { useState, useEffect } from 'react';
import './index.css';
import config from './config';
import { validateAndNormalizeURL } from './utils/urlValidation';

function App() {
  const [url, setUrl] = useState('');
  const [loading, setLoading] = useState(false);
  const [results, setResults] = useState(null);
  const [error, setError] = useState('');
  const [isAsync, setIsAsync] = useState(false);
  const [analysisJobs, setAnalysisJobs] = useState([]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    const validation = validateAndNormalizeURL(url.trim());
    if (!validation.isValid) {
      setError(validation.error);
      return;
    }

    setLoading(true);
    setError('');
    setResults(null);

    try {
      const response = await fetch(`${config.api.baseUrl}${config.api.endpoints.analyze}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ 
          url: validation.normalizedUrl,
          async: isAsync,
          priority: 1
        }),
      });

      const data = await response.json();


      if (!response.ok) {
        setError(`HTTP ${response.status}: ${data.error || 'Unknown error'}`);
      } else if (data.error) {
        setError(`${data.error}${data.statusCode ? ` (Status: ${data.statusCode})` : ''}`);
      } else {
        if (isAsync) {
          const newJob = {
            ...data,
            submittedAt: new Date().toLocaleTimeString()
          };
          setAnalysisJobs(prev => [newJob, ...prev]);
          setUrl('');
          
          if (data.status === 'pending') {
            setTimeout(() => startPolling(data.analysis_id), config.polling.interval);
          }
        } else {
          setResults(data);
        }
      }
    } catch (err) {
      setError('Failed to analyze the webpage. Please check if the backend server is running.');
    } finally {
      setLoading(false);
    }
  };

  const pollJobStatus = async (analysisId) => {
    		try {
			const response = await fetch(`${config.api.baseUrl}${config.api.endpoints.analysis}/${analysisId}`);
			const data = await response.json();
			
			if (response.ok) {
				setAnalysisJobs(prev => 
					prev.map(job => 
						job.analysis_id === analysisId ? { ...job, ...data } : job
					)
				);
				return data;
			}
    } catch (err) {
      console.error('Failed to poll job status:', err);
    }
    return null;
  };

  const handleViewDetails = async (job) => {
          if (job.status === 'pending' || job.status === 'processing') {
        const updatedJob = await pollJobStatus(job.analysis_id);
      if (updatedJob && updatedJob.status === 'completed') {
        setResults(updatedJob);
      }
    } else if (job.status === 'completed') {
      setResults(job);
    }
  };

  const startPolling = (analysisId) => {
    let attempts = 0;
    const interval = setInterval(async () => {
      attempts++;
      const job = analysisJobs.find(j => j.analysis_id === analysisId);
      if (job && (job.status === 'pending' || job.status === 'processing')) {
        const updatedJob = await pollJobStatus(analysisId);
        if (updatedJob && (updatedJob.status === 'completed' || updatedJob.status === 'failed')) {
          clearInterval(interval);
        }
      } else {
        clearInterval(interval);
      }
      
      if (attempts >= config.polling.maxAttempts) {
        clearInterval(interval);
      }
    }, config.polling.interval);

    return interval;
  };

  const renderHeadings = (headings) => {
    const headingLevels = ['h1', 'h2', 'h3', 'h4', 'h5', 'h6'];
    const safeHeadings = headings || {};
    return (
      <div className="headings-grid">
        {headingLevels.map(level => (
          <div key={level} className="heading-item">
            <div className="level">{level}</div>
            <div className="count">{safeHeadings[level] || 0}</div>
          </div>
        ))}
      </div>
    );
  };

  const renderLinks = (links) => {
    const safeLinks = links || {};
    return (
      <div className="links-grid">
        <div className="link-item">
          <div className="type">Internal</div>
          <div className="count">{safeLinks.internal || 0}</div>
        </div>
        <div className="link-item">
          <div className="type">External</div>
          <div className="count">{safeLinks.external || 0}</div>
        </div>
        <div className="link-item">
          <div className="type">Inaccessible</div>
          <div className="count">{safeLinks.inaccessible || 0}</div>
        </div>
      </div>
    );
  };

  return (
    <div className="container">
      <div className="header">
        <h1>Web Page Analyzer</h1>
        <p>Analyze web pages with synchronous or asynchronous processing</p>
      </div>

      <form onSubmit={handleSubmit} className="analyzer-form">
        <div className="form-group">
          <label htmlFor="url">Enter URL or domain to analyze:</label>
          <input
            type="text"
            id="url"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder="example.com, svgtopng.com, https://github.com"
            required
          />
        </div>
        
        <div className="form-group">
          <div className="toggle-container">
            <label className="toggle-label">
              <input
                type="checkbox"
                checked={isAsync}
                onChange={(e) => setIsAsync(e.target.checked)}
                className="toggle-input"
              />
              <span className="toggle-slider"></span>
              <span className="toggle-text">
                {isAsync ? 'Asynchronous Processing' : 'Synchronous Processing'}
              </span>
            </label>
            <p className="toggle-description">
              {isAsync 
                ? 'Submit multiple URLs for background processing. View results when ready.'
                : 'Get immediate results. Perfect for single URL analysis.'
              }
            </p>
          </div>
        </div>
        
        <button type="submit" disabled={loading} className="analyze-btn">
          {loading ? 'Analyzing...' : (isAsync ? 'Submit for Processing' : 'Analyze Page')}
        </button>
      </form>

      {loading && (
        <div className="loading">
          Analyzing webpage, please wait...
        </div>
      )}

      {error && (
        <div className="error">
          <strong>Error:</strong> {error}
        </div>
      )}

      {isAsync && analysisJobs.length > 0 && (
        <div className="async-jobs">
          <h2>Analysis Jobs</h2>
          <div className="jobs-grid">
            {analysisJobs.map((job) => (
              <div key={job.id} className={`job-card ${job.status}`}>
                <div className="job-header">
                  <h3>{job.url}</h3>
                  <span className={`status-badge ${job.status}`}>
                    {job.status}
                  </span>
                </div>
                <div className="job-meta">
                  <p><strong>Submitted:</strong> {job.submittedAt}</p>
                  <p><strong>Job ID:</strong> {job.id}</p>
                  <p><strong>Analysis ID:</strong> {job.analysis_id}</p>
                </div>
                <div className="job-actions">
                  {job.status === 'completed' ? (
                    <button 
                      className="view-details-btn"
                      onClick={() => handleViewDetails(job)}
                    >
                      View Details
                    </button>
                  ) : job.status === 'failed' ? (
                    <span className="error-text">Analysis Failed</span>
                  ) : (
                    <div className="processing-indicator">
                      <span>Processing...</span>
                      <button 
                        className="refresh-btn"
                        onClick={() => pollJobStatus(job.analysis_id)}
                      >
                        Refresh
                      </button>
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {results && (
        <div className="results">
          <h2>Analysis Results</h2>
  
          <div className="analysis-meta">
            <p><strong>URL:</strong> {results.url}</p>
            <p><strong>Status:</strong> {results.status}</p>
            <p><strong>Analysis ID:</strong> {results.id}</p>
          </div>
          

          
          {results.result && (
            <>
              <div className="result-section">
                <h3>HTML Version</h3>
                <div className="result-value">{results.result?.html_version || 'Unknown'}</div>
              </div>

              <div className="result-section">
                <h3>Page Title</h3>
                <div className="result-value">{results.result?.title || 'No title found'}</div>
              </div>

              <div className="result-section">
                <h3>Headings Distribution</h3>
                {renderHeadings(results.result?.headings)}
              </div>

              <div className="result-section">
                <h3>Links Analysis</h3>
                {renderLinks(results.result?.links)}
              </div>

              <div className="result-section">
                <h3>Login Form Detection</h3>
                <div className={`login-form-status ${results.result?.has_login_form ? 'has-login' : 'no-login'}`}>
                  {results.result?.has_login_form ? '✓ Login form detected' : '✗ No login form found'}
                </div>
              </div>

              <div className="result-section">
                <h3>Performance Metrics</h3>
                <div className="metrics-grid">
                  <div className="metric-item">
                    <div className="metric-label">Load Time</div>
                    <div className="metric-value">{((results.result?.load_time || 0) / 1000000).toFixed(2)} ms</div>
                  </div>
                  <div className="metric-item">
                    <div className="metric-label">Content Length</div>
                    <div className="metric-value">{((results.result?.content_length || 0) / 1024).toFixed(2)} KB</div>
                  </div>
                  <div className="metric-item">
                    <div className="metric-label">Status Code</div>
                    <div className="metric-value">{results.result?.status_code || 'Unknown'}</div>
                  </div>
                </div>
              </div>

              {results.result?.links?.external_hosts && results.result.links.external_hosts.length > 0 && (
                <div className="result-section">
                  <h3>External Hosts</h3>
                  <div className="external-hosts">
                    {results.result.links.external_hosts.slice(0, 10).map((host, index) => (
                      <span key={index} className="host-tag">{host}</span>
                    ))}
                    {results.result.links.external_hosts.length > 10 && (
                      <span className="more-hosts">... and {results.result.links.external_hosts.length - 10} more</span>
                    )}
                  </div>
                </div>
              )}
            </>
          )}
        </div>
      )}
    </div>
  );
}

export default App;
