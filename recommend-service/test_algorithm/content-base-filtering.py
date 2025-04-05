import pandas as pd
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics.pairwise import cosine_similarity
# Dữ liệu giả lập: danh sách phim và thể loại
movies = {
    'title': ['The Dark Knight', 'Inception', 'Titanic', 'The Avengers', 'The Notebook'],
    'genres': ['action crime drama', 'action sci-fi thriller', 'romance drama', 'action superhero', 'romance drama'],
    'votes': [950000, 870000, 600000, 920000, 450000]
}

# Phim người dùng đã thích
user_likes = ['The Dark Knight', 'Inception']

# Tạo DataFrame
df_movies = pd.DataFrame(movies)
print("Danh sách phim:")
print(df_movies)

# Khởi tạo TfidfVectorizer
tfidf = TfidfVectorizer(stop_words='english')  # Loại bỏ từ không quan trọng như "and", "the"

# Chuyển đổi cột 'genres' thành ma trận TF-IDF
tfidf_matrix = tfidf.fit_transform(df_movies['genres'])

# Tạo vector sở thích của người dùng
user_profile = tfidf_matrix[df_movies[df_movies['title'].isin(user_likes)].index].mean(axis=0)

# Tính cosine similarity giữa tất cả các phim
cosine_sim = cosine_similarity(tfidf_matrix, tfidf_matrix)
print("cosine_sim:",cosine_sim)

# Tính độ tương đồng giữa user_profile và tất cả phim
user_sim_scores = cosine_similarity(user_profile, tfidf_matrix)
print("user_sim_scores:",user_sim_scores)
def get_recommentdations_for_user(cosine_sim=user_sim_scores, df=df_movies, top_n=3):
    
    sim_scores = list(enumerate(cosine_sim[0]))
    print("sim_scores:",sim_scores)
    # Thêm tính toán trọng số votes vào cho cosine_sim
    max_votes = df['votes'].max()
    for index in range(len(sim_scores)):
        vote_weight = df["votes"][index]/max_votes
        combine_score = sim_scores[index][1] * 0.7 + vote_weight *0.3
        sim_scores[index]= (index,combine_score)
    
    # Sắp xếp theo độ tương đồng giảm dần
    sim_scores = sorted(sim_scores, key=lambda x: x[1], reverse=True)
    # Lấy 3 phim tương tự nhất (bỏ phim đầu tiên vì nó là chính phim đó)
    sim_scores = sim_scores[0:top_n]
    movie_indices = [i[0] for i in sim_scores]
    return df_movies['title'].iloc[movie_indices]

print("Phim gợi ý cho user:")
print(get_recommentdations_for_user())

# def get_recommendations(title, cosine_sim=user_sim_scores, df=df_movies, top_n=3):
#     # Tìm chỉ số (index) của phim trong DataFrame
#     # Khởi tạo biến để lưu chỉ số
#     idx = None
    
#     # Duyệt qua từng hàng trong df_movies
#     for i in range(len(df_movies)):
#         # Lấy tên phim ở hàng hiện tại
#         current_title = df_movies['title'][i]
        
#         # So sánh tên phim hiện tại với tên phim cần tìm
#         if current_title == title:
#             idx = i  # Lưu chỉ số nếu tìm thấy
#             break  # Thoát vòng lặp khi đã tìm thấy
    
#     # Lấy danh sách độ tương đồng của phim đó với tất cả phim khác
#     sim_scores = list(enumerate(cosine_sim[idx]))

#     # Bước 3: Kết hợp vote vào độ tương đồng
#     max_votes = df['votes'].max()  # Lấy số vote cao nhất để chuẩn hóa
#     for i in range(len(sim_scores)):
#             movie_idx = sim_scores[i][0]  # Chỉ số phim
#             sim_score = sim_scores[i][1]  # Độ tương đồng gốc
#             vote_weight = df['votes'][movie_idx] / max_votes  # Trọng số vote (0 đến 1)
#             # Cập nhật điểm số: 70% dựa trên thể loại, 30% dựa trên vote
#             combined_score = 0.5 * sim_score + 0.5 * vote_weight
#             sim_scores[i] = (movie_idx, combined_score)

#     # Sắp xếp theo độ tương đồng giảm dần
#     sim_scores = sorted(sim_scores, key=lambda x: x[1], reverse=True)
    
#     # Lấy 3 phim tương tự nhất (bỏ phim đầu tiên vì nó là chính phim đó)
#     sim_scores = sim_scores[1:top_n+1]
    
#     # Lấy chỉ số của các phim tương tự
#     movie_indices = [i[0] for i in sim_scores]
    
#     # Trả về tên các phim gợi ý
#     return df_movies['title'].iloc[movie_indices]

# # Thử gợi ý cho phim 'The Dark Knight'
# print("Phim gợi ý cho 'Inception':")
# print(get_recommendations('Inception'))