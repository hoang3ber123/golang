import pandas as pd
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics.pairwise import cosine_similarity
from sklearn.linear_model import LinearRegression
from sklearn.preprocessing import StandardScaler
import numpy as np

# Dữ liệu phim (thêm price, runtime, year)
data = {
    'user_id': [12, 15, 18, 21, 14, 22, 19, 11, 17, 23, 16, 13, 20, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40],
    'click-time': [2, 5, 3, 7, 6, 4, 2, 3, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,17, 18, 19, 20, 21, 22, 23, 24, 25, 26],
    'click-type': [0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 1, 1, 0, 1, 0, 1, 0, 1,1, 1, 1, 1, 0, 1, 0, 1, 0, 1],
    'film title': ['Halloween II', 'Inception', 'The Conjuring', 'Interstellar', 'The Dark Knight', 
                   'Parasite', 'A Quiet Place', 'Avengers: Endgame', 'Joker', 'Get Out', 
                   'The Matrix', 'Titanic', 'The Shining', 'Pulp Fiction', 'Gladiator', 
                   'The Godfather', "Schindler's List", 'Fight Club', 'The Silence of the Lambs', 'Mad Max: Fury Road','The Revenant', 'Django Unchained', 'The Departed', 'The Prestige', 'Whiplash', 
                   'The Grand Budapest Hotel', 'Blade Runner 2049', 'No Country for Old Men', 'Black Swan', 'Se7en'],
    'film id': ['film-01', 'film-02', 'film-03', 'film-04', 'film-05', 'film-06', 'film-07', 'film-08', 'film-09', 'film-10',
                'film-11', 'film-12', 'film-13', 'film-14', 'film-15', 'film-16', 'film-17', 'film-18', 'film-19', 'film-20','film-21', 'film-22', 'film-23', 'film-24', 'film-25', 'film-26', 'film-27', 'film-28', 'film-29', 'film-30'],
    'genres title': ['action horror', 'sci-fi thriller', 'horror mystery', 'sci-fi drama', 'action crime', 
                     'thriller drama', 'horror thriller', 'action sci-fi', 'drama thriller', 'horror mystery', 
                     'sci-fi action', 'romance drama', 'horror thriller', 'crime drama', 'action drama', 
                     'crime drama', 'historical drama', 'drama thriller', 'crime thriller', 'action sci-fi','adventure drama', 'western drama', 'crime thriller', 'mystery drama', 'music drama', 
                     'comedy drama', 'sci-fi thriller', 'crime drama', 'psychological thriller', 'crime mystery']
}

df = pd.DataFrame(data)
df_movies = df[['film id', 'film title', 'genres title']].drop_duplicates().reset_index(drop=True)
df_movies.columns = ['film_id', 'film_title', 'genres']

# Thêm votes, price, runtime, year giả lập
df_movies['votes'] = np.random.randint(40, 100, size=len(df_movies))
df_movies['price'] = np.random.uniform(5, 15, size=len(df_movies))
df_movies['runtime'] = np.random.randint(90, 200, size=len(df_movies))
df_movies['year'] = np.random.randint(1990, 2023, size=len(df_movies))
# Vector hóa thể loại
tfidf = TfidfVectorizer(stop_words='english')
tfidf_matrix = tfidf.fit_transform(df_movies['genres'])

# Chọn user_id = 15
user_id = 15
user_interactions = df[df['user_id'] == user_id]
user_films = user_interactions['film title']
film_indices = df_movies[df_movies['film_title'].isin(user_films)].index
user_vectors = tfidf_matrix[film_indices]
user_profile = user_vectors.mean(axis=0)
user_profile = np.asarray(user_profile)

cosine_sim = cosine_similarity(user_profile.reshape(1, -1), tfidf_matrix)[0]

# Chuẩn hóa các đặc trưng số
votes_normalized = (df_movies['votes'] - df_movies['votes'].min()) / (df_movies['votes'].max() - df_movies['votes'].min())
price_normalized = (df_movies['price'] - df_movies['price'].min()) / (df_movies['price'].max() - df_movies['price'].min())
runtime_normalized = (df_movies['runtime'] - df_movies['runtime'].min()) / (df_movies['runtime'].max() - df_movies['runtime'].min())
year_normalized = (df_movies['year'] - df_movies['year'].min()) / (df_movies['year'].max() - df_movies['year'].min())
# Tạo nhãn dựa trên click-time
labels = []
for title in df_movies['film_title']:
    if title in user_interactions['film title'].values:
        click_time = user_interactions[user_interactions['film title'] == title]['click-time'].values[0]
        labels.append(click_time)
    else:
        labels.append(0)

training_data = pd.DataFrame({
    'sim_score': cosine_sim,
    'vote_weight': votes_normalized,
    'price_weight': price_normalized,
    'runtime_weight': runtime_normalized,
    'year_weight': year_normalized,
    'label': labels
})

# Huấn luyện mô hình
X = training_data[['sim_score', 'vote_weight', 'price_weight', 'runtime_weight', 'year_weight']]
y = training_data['label']
model = LinearRegression()
model.fit(X, y)

# Lấy trọng số
weights = model.coef_
intercept = model.intercept_
print("\nTrọng số đã học (dữ liệu lớn):")
print(f"Trọng số cho sim_score: {weights[0]:.3f}")
print(f"Trọng số cho vote_weight: {weights[1]:.3f}")
print(f"Trọng số cho price_weight: {weights[2]:.3f}")
print(f"Trọng số cho runtime_weight: {weights[3]:.3f}")
print(f"Trọng số cho year_weight: {weights[4]:.3f}")
print(f"Hệ số chặn: {intercept:.3f}")

# Gợi ý
combined_scores = (weights[0] * cosine_sim +
                  weights[1] * votes_normalized +
                  weights[2] * price_normalized +
                  weights[3] * runtime_normalized +
                  weights[4] * year_normalized +
                  intercept)
top_indices = combined_scores.argsort()[::-1]
recommended_indices = [idx for idx in top_indices if df_movies['film_title'].iloc[idx] not in user_films.values][:3]
recommended_movies = df_movies['film_title'].iloc[recommended_indices]
print("\nPhim gợi ý cho user_id 15:")
print(recommended_movies)