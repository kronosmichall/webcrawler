from pymongo import MongoClient
from sklearn.feature_extraction.text import TfidfVectorizer
from nltk.corpus import stopwords
from nltk.tokenize import word_tokenize
import nltk
import string

nltk.download("punkt")
nltk.download("stopwords")

client = MongoClient("mongodb://localhost:27017")
db = client["web"]
collection = db["websites"]

docs = list(collection.find({}, {"title": 1, "body": 1, "url": 1}))

urls = []
texts = []

for doc in docs:
    url = doc.get("url")
    title = doc.get("title")
    body = doc.get("body")
    full_text = f"{title} {body}".lower()

    urls.append(url)
    texts.append(full_text)

vectorizer = TfidfVectorizer(stop_words="english", max_features=1000)
X = vectorizer.fit_transform(texts)
feature_names = vectorizer.get_feature_names_out()

url_keywords = {}

for i, url in enumerate(url):
    row = X[i].tocoo()
    top_indices = row.col[row.data.argsort()[::-1][:5]]
    keywords = [feature_names[idx] for idx in top_indices]
    url_keywords[url] = keywords

for url, keywords in url_keywords:
    print(f"{url}: {keywords}")
