// base url from fabric explorer
const API_BASE = process.env.API_BASE || 'http://localhost:8000';
const DEFAULT_CHAINCODE = 'mycc';
const DEFAULT_USER = 'user1';

export const rejectErrors = res => {
  const { status } = res;
  if (status >= 200 && status < 300) {
    return res;
  }

  return Promise.reject({ message: res.statusText, status });
};

export const fetchJson = (url, options = {}, base = API_BASE) =>
  fetch(/^(?:https?)?:\/\//.test(url) ? url : base + url, options)
    .then(rejectErrors)
    // default return empty json when no content
    .then(res => res.json());

export const getArguments = args =>
  args ? args.map(value => '&arguments[]=' + value).join('') : '';

export const query = (fcn, ...args) =>
  fetchJson(
    `/api/query?chaincode=${DEFAULT_CHAINCODE}&method=${fcn}&user=${DEFAULT_USER}` +
      getArguments(args)
  );

export const invoke = (fcn, ...args) =>
  fetchJson(
    `/api/invoke?chaincode=${DEFAULT_CHAINCODE}&method=${fcn}&user=${DEFAULT_USER}` +
      getArguments(args)
  );
