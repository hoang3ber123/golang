import grpc
from concurrent import futures
import time
import sys
import os
import pandas as pd
# Thêm thư mục chứa 'proto-golang' vào sys.path
sys.path.append(os.path.join(os.path.dirname(__file__), 'proto-golang'))

# Import các module từ 'proto-golang/recommend'
from recommend import recommend_pb2_grpc
from recommend import recommend_pb2
from services.category_service import predict_categories
from services.recommend_service import recommend_products


# Implement service RecommendService
class RecommendServiceServicer(recommend_pb2_grpc.RecommendServiceServicer):
    
    def GetRecommendCategoryIDs(self, request, context):
        # Lấy dữ liệu từ request
        query = request.query  # Chuỗi truy vấn
        categories = request.categories  # Danh sách các Category từ request
        # Gọi hàm predict_categories từ category_service.py
        category_ids = predict_categories(query, categories, threshold=0.60, rank=10)
        
        # Tạo response theo định nghĩa protobuf
        response = recommend_pb2.GetRecommendCategoryIDsResponse()
        response.category_ids.extend(category_ids)  # Gán danh sách category_ids vào response
        
        # Trả về response
        return response
    
    def GetRecommendProductIDs(self, request, context):
        # Chuyển đổi dữ liệu từ gRPC request thành DataFrame
        products = []
        for p in request.products:
            products.append({
                'id': p.id,
                'title': p.title,
                'categories': list(p.categories),  # Repeated field thành list
                'created_at': p.created_at,
                'pricing': p.pricing
            })
        products_df = pd.DataFrame(products) if products else pd.DataFrame(columns=['id', 'title', 'categories', 'created_at', 'pricing'])

        click_details = [{'product_id': c.product_id, 'click_time': c.click_time} for c in request.click_details]
        click_df = pd.DataFrame(click_details) if click_details else pd.DataFrame(columns=['product_id', 'click_time'])

        view_products = [{'product_id': v.product_id, 'view_time': v.view_time} for v in request.view_products]
        view_df = pd.DataFrame(view_products) if view_products else pd.DataFrame(columns=['product_id', 'view_time'])

        bought_products = [{'product_id': b} for b in request.bought_products]
        bought_df = pd.DataFrame(bought_products) if bought_products else pd.DataFrame(columns=['product_id'])

        # Gọi hàm recommend_products
        try:
            product_ids = recommend_products(products_df, click_df, view_df, bought_df)
        except Exception as e:
            context.set_details(f"Error in recommend_products: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            return recommend_pb2.GetRecommendProductIDsResponse()

        # Tạo response
        response = recommend_pb2.GetRecommendProductIDsResponse()
        response.product_ids.extend(product_ids)
        return response

def serve():
    # Khởi tạo gRPC server
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    
    # Thêm service vào server
    recommend_pb2_grpc.add_RecommendServiceServicer_to_server(
        RecommendServiceServicer(), server
    )
    
    # Lắng nghe trên port 50054
    server.add_insecure_port('[::]:50054')
    print("gRPC server is starting on port 50054...")
    
    # Khởi động server
    server.start()
    
    # Giữ server chạy
    try:
        while True:
            time.sleep(86400)  # Chạy mãi mãi, sleep 1 ngày
    except KeyboardInterrupt:
        server.stop(0)
        print("gRPC server stopped.")

if __name__ == '__main__':
    serve()