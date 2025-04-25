import pandas as pd
import numpy as np
from sklearn.preprocessing import MultiLabelBinarizer
from sklearn.metrics.pairwise import cosine_similarity
from sklearn.linear_model import LogisticRegression

# Hàm recommend_products dùng DataFrame
def recommend_products(products_df, click_df, view_df, bought_df):
    # Biểu diễn đặc trưng
    mlb = MultiLabelBinarizer()
    categories_encoded = mlb.fit_transform(products_df['categories'])
    products_features = pd.DataFrame(categories_encoded, columns=mlb.classes_, index=products_df['id'])
    products_features['pricing'] = products_df.set_index('id')['pricing'] / products_df['pricing'].max()

    # Lịch sử tương tác
    user_clicks = click_df['product_id'].value_counts().to_dict()
    user_views = view_df['product_id'].value_counts().to_dict()
    user_bought = bought_df['product_id'].to_list()  # Danh sách sản phẩm đã mua

    # Tạo user_interactions với tất cả sản phẩm từ products_df
    user_interactions = pd.DataFrame({
        'product_id': products_df['id'],
        'click_count': [user_clicks.get(pid, 0) for pid in products_df['id']],
        'view_count': [user_views.get(pid, 0) for pid in products_df['id']],
        'bought': [1 if pid in user_bought else 0 for pid in products_df['id']]
    })

    # Kiểm tra số lớp trong bought
    X = user_interactions[['click_count', 'view_count']]
    y = user_interactions['bought']
    if len(y.unique()) < 2:  # Nếu chỉ có 1 lớp
        w_click_trained, w_view_trained = 0.3, 0.5
        w_bought_trained = 1.0
    else:
        model = LogisticRegression(solver='lbfgs')
        model.fit(X, y)
        w_click_trained, w_view_trained = model.coef_[0]
        w_bought_trained = 1.0

    # Tính profile người dùng
    user_profile = np.zeros(products_features.shape[1])
    for prod_id in products_features.index:
        click_count = user_clicks.get(prod_id, 0)
        view_count = user_views.get(prod_id, 0)
        bought = 1 if prod_id in user_bought else 0
        interactions = (w_click_trained * click_count +
                        w_view_trained * view_count +
                        w_bought_trained * bought)
        if interactions > 0:
            user_profile += interactions * products_features.loc[prod_id].values
    user_profile = user_profile / (np.linalg.norm(user_profile) + 1e-6)

    # Tính tương đồng
    similarities = cosine_similarity([user_profile], products_features.values)[0]
    products_df['similarity'] = similarities

    # Gợi ý top 10 sản phẩm chưa mua
    recommended_products = products_df[~products_df['id'].isin(user_bought)]
    top_10 = recommended_products.sort_values('similarity', ascending=False).head(10)
    return top_10['id'].tolist()  # Trả về danh sách ID


