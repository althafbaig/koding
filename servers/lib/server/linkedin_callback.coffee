{
  redirectOauth
  saveOauthToSession
}                  = require "./helpers"
{linkedin}         = KONFIG
http               = require "https"
{parseString}      = require "xml2js"
querystring        = require "querystring"
url                = require "url"
provider           = "linkedin"

module.exports = (req, res) ->
  access_token  = null
  expires_in    = null
  {code}        = req.query
  {clientId}    = req.cookies
  {
    client_id
    client_secret
    redirect_uri
  }              = linkedin

  unless code
    redirectOauth res, {provider}, "No code in query"
    return

  # Get user info with access token
  fetchUserInfo = (userInfoResp)->
    rawResp = ""
    userInfoResp.on "data", (chunk) -> rawResp += chunk
    userInfoResp.on "end", ->
      try
        parseString rawResp, (err, result) ->
          if err
            redirectOauth res, {provider}, "Error parsing user info"
            return

          try
            profileUrl = result.person['site-standard-profile-request'][0].url[0]
            {id} = querystring.decode(url.parse(profileUrl).query)
          catch e
            redirectOauth res, {provider}, "Error parsing user id"
            return

          linkedInResp =
            foreignId : id
            token     : access_token
            expires   : expires_in
            profile   : result.person

          saveOauthToSession linkedInResp, clientId, provider, (err)->
            if err
              redirectOauth res, {provider}, err
              return

            redirectOauth res, {provider}, null
      catch e
        redirectOauth res, {provider}, "Error parsing user info"

  # Get access token with code
  authorizeUser = (authUserResp)->
    rawResp = ""
    authUserResp.on "data", (chunk) -> rawResp += chunk
    authUserResp.on "end", ->
      try
        tokenInfo = JSON.parse rawResp
      catch e
        redirectOauth res, {provider}, "Error getting access token"

      {access_token, expires_in} = tokenInfo
      if access_token
        options =
          host   : "api.linkedin.com"
          path   : "/v1/people/~?oauth2_access_token=#{access_token}"
          method : "GET"
        re = http.request options, fetchUserInfo
        re.end()
      else
        redirectOauth res, {provider}, "No access token"

  path  = "/uas/oauth2/accessToken?"
  path += "grant_type=authorization_code&"
  path += "code=#{code}&"
  path += "redirect_uri=#{redirect_uri}&"
  path += "client_id=#{client_id}&"
  path += "client_secret=#{client_secret}"

  options   =
    host    : "www.linkedin.com"
    path    : path
    method  : "GET"

  r = http.request options, authorizeUser
  r.end()

  r.on 'error', (e)-> console.log 'problem with request: ' + e.message
