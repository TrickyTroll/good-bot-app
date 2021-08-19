import json
import requests
import os
import sqlite3
from flask_login import (
    LoginManager,
    current_user,
    login_required,
    login_user,
    logout_user,
)
from flask import (
    Flask,
    render_template,
    flash,
    redirect,
    request,
    session,
    abort,
    url_for,
)
from oauthlib.oauth2 import WebApplicationClient
from werkzeug.utils import secure_filename
from markupsafe import escape
from db import init_db_command
from user import User

GOOGLE_CLIENT_ID = os.environ.get("GOOGLE_CLIENT_ID", None)
GOOGLE_CLIENT_SECRET = os.environ.get("GOOGLE_CLIENT_SECRET", None)
GOOGLE_DISCOVERY_URL = (
    "https://accounts.google.com/.well-known/openid-configuration"
)

app = Flask(__name__)

app.secret_key = os.environ.get("SECRET_KEY") or os.urandom(24)
app.config['MAX_CONTENT_LENGTH'] = 1024 * 1024
app.config['UPLOAD_EXTENSIONS'] = ['.txt', '.yaml' '.yml']

login_manager = LoginManager()
login_manager.init_app(app)

try:
    init_db_command()
except sqlite3.OperationalError:
    print("DB is already set up.")
    pass

client = WebApplicationClient(GOOGLE_CLIENT_ID)

def get_google_provider_cfg():
    return requests.get(GOOGLE_DISCOVERY_URL).json()

@login_manager.user_loader
def load_user(user_id):
    return User.get(user_id)

@app.route('/')
def index():
    return render_template('index.html')

@app.route("/login")
def login():
    # Find out what URL to hit for Google login
    if current_user.is_authenticated:
        flash('You are already logged in')
        return redirect( url_for('index') )
    google_provider_cfg = get_google_provider_cfg()
    authorization_endpoint = google_provider_cfg["authorization_endpoint"]

    # Use library to construct the request for Google login and provide
    # scopes that let you retrieve user's profile from Google
    request_uri = client.prepare_request_uri(
        authorization_endpoint,
        redirect_uri=request.base_url + "/callback",
        scope=["openid", "email", "profile"],
    )
    return redirect(request_uri)


@app.route("/login/callback")
def callback():
    # Get authorization code Google sent back to you
    code = request.args.get("code")
    google_provider_cfg = get_google_provider_cfg()
    token_endpoint = google_provider_cfg["token_endpoint"]
    token_url, headers, body = client.prepare_token_request(
        token_endpoint,
        authorization_response=request.url,
        redirect_url=request.base_url,
        code=code
    )
    token_response = requests.post(
        token_url,
        headers=headers,
        data=body,
        auth=(GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET),
    )
    # Parse the tokens!
    client.parse_request_body_response(json.dumps(token_response.json()))
    userinfo_endpoint = google_provider_cfg["userinfo_endpoint"]
    uri, headers, body = client.add_token(userinfo_endpoint)
    userinfo_response = requests.get(uri, headers=headers, data=body)
    if userinfo_response.json().get("email_verified"):
        unique_id = userinfo_response.json()["sub"]
        users_email = userinfo_response.json()["email"]
        picture = userinfo_response.json()["picture"]
        users_name = userinfo_response.json()["given_name"]
    else:
        return "User email not available or not verified by Google.", 400
    user = User(
        id_=unique_id, name=users_name, email=users_email, profile_pic=picture
    )

    # Doesn't exist? Add it to the database.
    if not User.get(unique_id):
        User.create(unique_id, users_name, users_email, picture)

    # Begin user session by logging the user in
    login_user(user)

    # Send user back to homepage
    flash('You were successfully logged in')
    return redirect(url_for("index"))

@app.route("/logout")
def logout():
    if current_user.is_authenticated:
        logout_user()
        flash('You were successfully logged out')
    else:
        flash('Your are not logged in')

    return redirect(url_for("index"))


@app.route('/post')
def user_post():
    if current_user.is_authenticated:
        return redirect(f'/post/{current_user.id}')

    else:
        flash('You need to be logged in to make a new post.')
        return redirect(url_for('index'))

@login_required
@app.route('/post/<user_id>')
def user_post_create(user_id):
    """
    Intermediate screen for the user to create a new post.
    """
    return render_template("post.html")

@login_required
@app.route('/post/record/<user_id>', methods=['POST'])
def user_post_new(user_id):
    # show the post with the given id, the id is an integer
    return f"{request.form['title']}\n{request.form['contents']}"


@app.route('/view/<user_id>/<title>')
def show_user_asciicasts(user_id, title):
    # show the user profile for that user
    return f'Hello {user_id}'

if __name__ == "__main__":
    app.run(ssl_context="adhoc")
