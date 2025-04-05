import pandas as pd
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics.pairwise import cosine_similarity
from sklearn.linear_model import LinearRegression
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import StandardScaler
import numpy as np

# Dữ liệu phim
movies = {
    'title': ['The Dark Knight', 'Inception', 'Titanic', 'The Avengers', 'The Notebook'],
    'genres': ['action crime drama', 'action sci-fi thriller', 'romance drama', 'action superhero', 'romance drama'],
    'votes': [950000, 870000, 600000, 920000, 450000]
}

# Tạo DataFrame
df_movies = pd.DataFrame(movies)
print("Danh sách phim:")
print(df_movies)

# Phim người dùng đã thích
user_likes = ['The Dark Knight', 'Inception']

# Khởi tạo TfidfVectorizer
tfidf = TfidfVectorizer(stop_words='english')

# Chuyển đổi cột 'genres' thành ma trận TF-IDF
tfidf_matrix = tfidf.fit_transform(df_movies['genres'])

# Tìm chỉ số của các phim người dùng đã thích
film_indices = df_movies[df_movies['title'].isin(user_likes)].index

# Lấy vector TF-IDF và tính hồ sơ người dùng
user_vectors = tfidf_matrix[film_indices]
user_profile = user_vectors.mean(axis=0)
user_profile = np.asarray(user_profile)

# Tính độ tương đồng giữa user_profile và tất cả phim
cosine_sim = cosine_similarity(user_profile.reshape(1, -1), tfidf_matrix)[0]

# Chuẩn hóa votes về [0, 1]
max_votes = df_movies['votes'].max()
vote_weights = df_movies['votes'] / max_votes

# Tạo dữ liệu giả lập: nhãn (1 nếu phim trong user_likes, 0 nếu không)
labels = [1 if title in user_likes else 0 for title in df_movies['title']]

# Tạo DataFrame chứa đặc trưng và nhãn
training_data = pd.DataFrame({
    'sim_score': cosine_sim,
    'vote_weight': vote_weights,
    'label': labels
})

print("\nDữ liệu huấn luyện:")
print(training_data)

# Tách đặc trưng và nhãn
X = training_data[['sim_score', 'vote_weight']]
y = training_data['label']

# Chia dữ liệu thành tập huấn luyện và kiểm tra (dù ở đây dữ liệu nhỏ)
X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)

# Khởi tạo và huấn luyện mô hình Linear Regression
model = LinearRegression()
model.fit(X_train, y_train)

# Lấy trọng số đã học
weights = model.coef_
print("\nTrọng số đã học:")
print(f"Trọng số cho sim_score: {weights[0]:.3f}")
print(f"Trọng số cho vote_weight: {weights[1]:.3f}")

# Hệ số chặn (intercept)
intercept = model.intercept_
print(f"Hệ số chặn: {intercept:.3f}")

# Tính điểm tổng hợp bằng trọng số đã học
combined_scores = weights[0] * cosine_sim + weights[1] * vote_weights + intercept

# Sắp xếp và gợi ý
top_indices = combined_scores.argsort()[::-1]
recommended_indices = [idx for idx in top_indices if df_movies['title'].iloc[idx] not in user_likes][:2]
recommended_movies = df_movies['title'].iloc[recommended_indices]
print("\nPhim gợi ý (dùng trọng số học máy):")
print(recommended_movies)