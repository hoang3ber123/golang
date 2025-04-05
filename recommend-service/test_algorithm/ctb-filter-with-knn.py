import pandas as pd
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.neighbors import NearestNeighbors
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
# print("Kích thước ma trận TF-IDF:", tfidf_matrix.shape)
# print("ma trận TF-IDF:", tfidf_matrix)

# Tìm chỉ số của các phim người dùng đã thích
film_indices = df_movies[df_movies['title'].isin(user_likes)].index

# Lấy vector TF-IDF của các phim
user_vectors = tfidf_matrix[film_indices]

# Tính vector trung bình (hồ sơ người dùng)
user_profile = user_vectors.mean(axis=0)
user_profile = np.asarray(user_profile)  # Chuyển thành numpy array để KNN xử lý
print("Hồ sơ người dùng (vector sở thích):")
print(user_profile)

# Khởi tạo NearestNeighbors
knn = NearestNeighbors(n_neighbors=5, metric='cosine')  # Tìm 3 phim gần nhất
knn.fit(tfidf_matrix)

# Chuyển user_profile thành ma trận để KNN xử lý
user_profile = user_profile.reshape(1, -1)

# Tìm các phim gần nhất với user_profile
distances, indices = knn.kneighbors(user_profile)
print("distances",distances)
print("indices",indices)
# Tính điểm số kết hợp (dựa trên khoảng cách và votes)
max_votes = df_movies['votes'].max()  # Chuẩn hóa votes
combined_scores = []
for i, idx in enumerate(indices[0]):
    if df_movies['title'].iloc[idx] in user_likes:
        continue  # Bỏ qua phim đã thích
    sim_score = 1 - distances[0][i]  # Chuyển khoảng cách thành độ tương đồng (1 - distance)
    vote_weight = df_movies['votes'].iloc[idx] / max_votes  # Trọng số votes
    combined_score = 0.7 * sim_score + 0.3 * vote_weight  # Kết hợp
    combined_scores.append((idx, combined_score))

# Sắp xếp theo điểm số kết hợp
combined_scores = sorted(combined_scores, key=lambda x: x[1], reverse=True)[:2]  # Lấy top 2
recommended_indices = [idx for idx, score in combined_scores]

# Trả về tên phim gợi ý
recommended_movies = df_movies['title'].iloc[recommended_indices]
print("\nPhim gợi ý (kết hợp votes):")
print(recommended_movies)