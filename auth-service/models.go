package main

type authRequestBody struct {
	UserID    string `json:"userID" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Operation string `json:"operation" binding:"required"`
}

type userPermissions struct {
	CanRead   bool
	CanWrite  bool
	CanDelete bool
}
