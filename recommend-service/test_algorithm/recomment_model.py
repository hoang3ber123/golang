import pandas as pd
import numpy as np
from sklearn.preprocessing import MultiLabelBinarizer
from sklearn.metrics.pairwise import cosine_similarity
from sklearn.linear_model import LogisticRegression

# Dữ liệu mẫu
products = [
    {'id': 'uuid1', 'title': 'Product 1', 'categories': ['backend', 'nodejs'], 'created_at': '2023-01-01', 'pricing': 100},
    {'id': 'uuid2', 'title': 'Product 2', 'categories': ['frontend', 'react'], 'created_at': '2023-02-01', 'pricing': 150},
    {'id': 'uuid3', 'title': 'Product 3', 'categories': ['backend', 'python'], 'created_at': '2023-03-01', 'pricing': 120},
    {'id': 'uuid4', 'title': 'Product 4', 'categories': ['backend', 'nodejs'], 'created_at': '2023-04-01', 'pricing': 110},
    {'id': 'uuid5', 'title': 'Product 5', 'categories': ['frontend', 'vue'], 'created_at': '2023-05-01', 'pricing': 130},
    {'id': 'uuid6', 'title': 'Product 6', 'categories': ['backend', 'java'], 'created_at': '2023-06-01', 'pricing': 140},
    {'id': 'uuid7', 'title': 'Product 7', 'categories': ['frontend', 'react'], 'created_at': '2023-07-01', 'pricing': 160},
    {'id': 'uuid8', 'title': 'Product 8', 'categories': ['backend', 'nodejs'], 'created_at': '2023-08-01', 'pricing': 125},
    {'id': 'uuid9', 'title': 'Product 9', 'categories': ['frontend', 'angular'], 'created_at': '2023-09-01', 'pricing': 135},
    {'id': 'uuid10', 'title': 'Product 10', 'categories': ['backend', 'python'], 'created_at': '2023-10-01', 'pricing': 115},
    {'id': 'uuid11', 'title': 'Product 11', 'categories': ['frontend', 'react'], 'created_at': '2023-11-01', 'pricing': 145},
    {'id': 'uuid12', 'title': 'Product 12', 'categories': ['backend', 'nodejs'], 'created_at': '2023-12-01', 'pricing': 150},
    {'id': 'uuid13', 'title': 'Product 13', 'categories': ['frontend', 'vue'], 'created_at': '2024-01-01', 'pricing': 125},
    {'id': 'uuid14', 'title': 'Product 14', 'categories': ['backend', 'java'], 'created_at': '2024-02-01', 'pricing': 130},
    {'id': 'uuid15', 'title': 'Product 15', 'categories': ['backend', 'python'], 'created_at': '2024-03-01', 'pricing': 120},
    {'id': 'uuid16', 'title': 'Product 16', 'categories': ['frontend', 'react'], 'created_at': '2024-04-01', 'pricing': 140},
    {'id': 'uuid17', 'title': 'Product 17', 'categories': ['backend', 'nodejs'], 'created_at': '2024-05-01', 'pricing': 110},
    {'id': 'uuid18', 'title': 'Product 18', 'categories': ['frontend', 'angular'], 'created_at': '2024-06-01', 'pricing': 135},
    {'id': 'uuid19', 'title': 'Product 19', 'categories': ['backend', 'python'], 'created_at': '2024-07-01', 'pricing': 125},
    {'id': 'uuid20', 'title': 'Product 20', 'categories': ['frontend', 'vue'], 'created_at': '2024-08-01', 'pricing': 145},
    {'id': 'uuid21', 'title': 'Product 21', 'categories': ['backend', 'java'], 'created_at': '2024-09-01', 'pricing': 150},
    {'id': 'uuid22', 'title': 'Product 22', 'categories': ['frontend', 'react'], 'created_at': '2024-10-01', 'pricing': 155},
    {'id': 'uuid23', 'title': 'Product 23', 'categories': ['backend', 'nodejs'], 'created_at': '2024-11-01', 'pricing': 130},
    {'id': 'uuid24', 'title': 'Product 24', 'categories': ['frontend', 'angular'], 'created_at': '2024-12-01', 'pricing': 125},
    {'id': 'uuid25', 'title': 'Product 25', 'categories': ['backend', 'python'], 'created_at': '2025-01-01', 'pricing': 120},
    {'id': 'uuid26', 'title': 'Product 26', 'categories': ['frontend', 'react'], 'created_at': '2025-02-01', 'pricing': 130},
    {'id': 'uuid27', 'title': 'Product 27', 'categories': ['backend', 'nodejs'], 'created_at': '2025-03-01', 'pricing': 140},
    {'id': 'uuid28', 'title': 'Product 28', 'categories': ['frontend', 'vue'], 'created_at': '2025-04-01', 'pricing': 145},
    {'id': 'uuid29', 'title': 'Product 29', 'categories': ['backend', 'java'], 'created_at': '2025-05-01', 'pricing': 135},
    {'id': 'uuid30', 'title': 'Product 30', 'categories': ['frontend', 'react'], 'created_at': '2025-06-01', 'pricing': 150},
    {'id': 'uuid31', 'title': 'Product 31', 'categories': ['backend', 'nodejs'], 'created_at': '2025-07-01', 'pricing': 120},
    {'id': 'uuid32', 'title': 'Product 32', 'categories': ['frontend', 'angular'], 'created_at': '2025-08-01', 'pricing': 125},
    {'id': 'uuid33', 'title': 'Product 33', 'categories': ['backend', 'python'], 'created_at': '2025-09-01', 'pricing': 115},
    {'id': 'uuid34', 'title': 'Product 34', 'categories': ['frontend', 'vue'], 'created_at': '2025-10-01', 'pricing': 130},
    {'id': 'uuid35', 'title': 'Product 35', 'categories': ['backend', 'java'], 'created_at': '2025-11-01', 'pricing': 140},
    {'id': 'uuid36', 'title': 'Product 36', 'categories': ['frontend', 'react'], 'created_at': '2025-12-01', 'pricing': 145},
    {'id': 'uuid37', 'title': 'Product 37', 'categories': ['backend', 'nodejs'], 'created_at': '2026-01-01', 'pricing': 125},
    {'id': 'uuid38', 'title': 'Product 38', 'categories': ['frontend', 'angular'], 'created_at': '2026-02-01', 'pricing': 135},
    {'id': 'uuid39', 'title': 'Product 39', 'categories': ['backend', 'python'], 'created_at': '2026-03-01', 'pricing': 140},
    {'id': 'uuid40', 'title': 'Product 40', 'categories': ['frontend', 'vue'], 'created_at': '2026-04-01', 'pricing': 150},
]


click_detail = [
    {'product_id': 'uuid1', 'click_time': '2023-01-02'},
    {'product_id': 'uuid2', 'click_time': '2023-01-03'},
    {'product_id': 'uuid3', 'click_time': '2023-01-04'},
]

view_product = [
    {'product_id': 'uuid1', 'view_time': '2023-01-02'},
    {'product_id': 'uuid2', 'view_time': '2023-01-03'},
    {'product_id': 'uuid3', 'view_time': '2023-01-04'},
]

bought_product = [
    {'product_id': 'uuid1'},
]

def recommend_products() -> pd.DataFrame:
    # Chuẩn bị dữ liệu
    products_df = pd.DataFrame(products)
    click_df = pd.DataFrame(click_detail)
    view_df = pd.DataFrame(view_product)
    bought_df = pd.DataFrame(bought_product)

    # Biểu diễn đặc trưng
    mlb = MultiLabelBinarizer()
    categories_encoded = mlb.fit_transform(products_df['categories'])
    products_features = pd.DataFrame(categories_encoded, columns=mlb.classes_, index=products_df['id'])
    products_features['pricing'] = products_df.set_index('id')['pricing'] / products_df['pricing'].max()

    # Lịch sử tương tác
    user_clicks = click_df['product_id'].value_counts().to_dict()
    user_views = view_df['product_id'].value_counts().to_dict()
    user_bought = bought_df['product_id'].value_counts().to_dict()

    # Train trọng số
    user_interactions = pd.DataFrame({
        'product_id': ['uuid1', 'uuid2', 'uuid3'],
        'click_count': [2, 0, 1],
        'view_count': [1, 3, 0],
        'bought': [1, 0, 0]
    })
    X = user_interactions[['click_count', 'view_count']]
    y = user_interactions['bought']
    model = LogisticRegression(solver='lbfgs')
    model.fit(X, y)
    w_click_trained, w_view_trained = model.coef_[0]
    w_bought_trained = 1.0

    # Tính profile người dùng
    user_profile = np.zeros(products_features.shape[1])
    for prod_id in products_features.index:
        click_count = user_clicks.get(prod_id, 0)
        view_count = user_views.get(prod_id, 0)
        bought = user_bought.get(prod_id, 0)
        interactions = (w_click_trained * click_count +
                        w_view_trained * view_count +
                        w_bought_trained * bought)
        if interactions > 0:
            user_profile += interactions * products_features.loc[prod_id].values
    user_profile = user_profile / (np.linalg.norm(user_profile) + 1e-6)

    # Tính tương đồng
    similarities = cosine_similarity([user_profile], products_features.values)[0]
    products_df['similarity'] = similarities

    # Gợi ý top 10
    recommended_products = products_df[~products_df['id'].isin(bought_df['product_id'])]
    top_10 = recommended_products.sort_values('similarity', ascending=False).head(10)
    return top_10[['id', 'title', 'similarity']]